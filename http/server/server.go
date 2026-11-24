// Derived from reqfleet/shibuya; modified for lib.
// Package server runs HTTP servers with context-owned graceful shutdown.
package server

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// StartServer serves HTTP, or HTTPS when both certificate files are supplied.
// Canceling ctx drains active requests for up to 30 seconds, then closes any
// remaining connections. A nil ctx disables automatic shutdown. Applications
// own signal handling; this package never installs signal handlers or exits the
// process. External Close/Shutdown and startup failures also release all workers.
func StartServer(server *http.Server, certFile, keyFile string, ctx context.Context) error {
	if (certFile == "") != (keyFile == "") {
		return errors.New("both TLS certificate and key files are required")
	}
	var done chan struct{}
	var shutdown chan error
	if ctx != nil && ctx.Done() != nil {
		done = make(chan struct{})
		shutdown = make(chan error, 1)
		go func() {
			select {
			case <-ctx.Done():
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				err := server.Shutdown(shutdownCtx)
				if err != nil {
					_ = server.Close()
				}
				shutdown <- err
			case <-done:
				shutdown <- nil
			}
		}()
	}

	var err error
	if certFile != "" {
		err = server.ListenAndServeTLS(certFile, keyFile)
	} else {
		err = server.ListenAndServe()
	}
	var shutdownErr error
	if shutdown != nil {
		close(done)
		shutdownErr = <-shutdown
	}
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return shutdownErr
}
