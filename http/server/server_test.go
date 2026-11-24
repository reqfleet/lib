package server_test

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/reqfleet/lib/http/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func wait[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case result := <-ch:
		return result
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for HTTP server")
		var zero T
		return zero
	}
}

func TestCancellationDrainsRequests(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ready := make(chan string, 1)
	entered := make(chan struct{})
	release := make(chan struct{})
	defer close(release)
	shuttingDown := make(chan struct{})
	srv := &http.Server{
		Addr: "127.0.0.1:0",
		BaseContext: func(listener net.Listener) context.Context {
			ready <- listener.Addr().String()
			return context.Background()
		},
		Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			close(entered)
			<-release
			_, _ = w.Write([]byte("finished"))
		}),
	}
	defer srv.Close()
	srv.RegisterOnShutdown(func() { close(shuttingDown) })
	done := make(chan error, 1)
	go func() { done <- server.StartServer(srv, "", "", ctx) }()
	address := wait(t, ready)
	response := make(chan *http.Response, 1)
	requestErr := make(chan error, 1)
	client := &http.Client{Timeout: 5 * time.Second}
	defer client.CloseIdleConnections()
	go func() {
		res, err := client.Get("http://" + address)
		response <- res
		requestErr <- err
	}()
	wait(t, entered)
	cancel()
	wait(t, shuttingDown)
	select {
	case err := <-done:
		t.Fatalf("server returned before draining its handler: %v", err)
	default:
	}
	release <- struct{}{}
	require.NoError(t, wait(t, requestErr))
	res := wait(t, response)
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Equal(t, "finished", string(body))
	require.NoError(t, wait(t, done))
}

func TestExternalClose(t *testing.T) {
	for _, ctx := range []context.Context{nil, context.Background()} {
		ready := make(chan string, 1)
		srv := &http.Server{Addr: "127.0.0.1:0", BaseContext: func(l net.Listener) context.Context {
			ready <- l.Addr().String()
			return context.Background()
		}}
		t.Cleanup(func() { _ = srv.Close() })
		done := make(chan error, 1)
		go func() { done <- server.StartServer(srv, "", "", ctx) }()
		wait(t, ready)
		require.NoError(t, srv.Close())
		require.NoError(t, wait(t, done))
	}
}

func TestStartupErrors(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()
	for _, tt := range []struct {
		name, address, cert, key string
	}{
		{"occupied address", listener.Addr().String(), "", ""},
		{"missing certificate", "127.0.0.1:0", "", "key.pem"},
		{"missing key", "127.0.0.1:0", "cert.pem", ""},
		{"invalid files", "127.0.0.1:0", filepath.Join(t.TempDir(), "missing.crt"), filepath.Join(t.TempDir(), "missing.key")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			srv := &http.Server{Addr: tt.address}
			defer srv.Close()
			done := make(chan error, 1)
			go func() { done <- server.StartServer(srv, tt.cert, tt.key, ctx) }()
			require.Error(t, wait(t, done))
		})
	}
}

func TestTLS(t *testing.T) {
	fixture := httptest.NewTLSServer(http.NotFoundHandler())
	defer fixture.Close()
	certificate := fixture.TLS.Certificates[0]
	key, err := x509.MarshalPKCS8PrivateKey(certificate.PrivateKey)
	require.NoError(t, err)
	dir := t.TempDir()
	certFile, keyFile := filepath.Join(dir, "cert.pem"), filepath.Join(dir, "key.pem")
	require.NoError(t, os.WriteFile(certFile, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certificate.Certificate[0]}), 0600))
	require.NoError(t, os.WriteFile(keyFile, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: key}), 0600))
	ready := make(chan string, 1)
	srv := &http.Server{
		Addr: "127.0.0.1:0",
		BaseContext: func(l net.Listener) context.Context {
			ready <- l.Addr().String()
			return context.Background()
		},
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.TLS == nil {
				http.Error(w, "TLS required", http.StatusBadRequest)
				return
			}
			_, _ = w.Write([]byte("secure"))
		}),
	}
	defer srv.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- server.StartServer(srv, certFile, keyFile, ctx) }()
	client := fixture.Client()
	client.Timeout = 5 * time.Second
	res, err := client.Get("https://" + wait(t, ready))
	require.NoError(t, err)
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "secure", string(body))
	cancel()
	require.NoError(t, wait(t, done))
}
