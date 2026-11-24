package auth

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
)

func TestAuthenticationContract(t *testing.T) {
	for _, tt := range []struct {
		name, path, bearer, required string
		status                       int
		body                         string
	}{
		{"missing", "/metrics", "", "secret", 400, "bearer header is empty\n"},
		{"wrong case", "/metrics", "bearer secret", "secret", 400, "token is invalid\n"},
		{"missing separator", "/metrics", "Bearer", "secret", 400, "token is invalid\n"},
		{"extra separator", "/metrics", "Bearer  secret", "secret", 400, "token is invalid\n"},
		{"trailing separator", "/metrics", "Bearer secret ", "secret", 400, "token is invalid\n"},
		{"wrong token", "/metrics", "Bearer wrong", "secret", 403, "forbidden\n"},
		{"empty token", "/metrics", "Bearer ", "secret", 403, "forbidden\n"},
		{"correct", "/metrics", "Bearer secret", "secret", 202, "accepted"},
		{"empty key and token", "/metrics", "Bearer ", "", 503, "bearer authentication is not configured\n"},
		{"empty key without credentials", "/metrics", "", "", 503, "bearer authentication is not configured\n"},
		{"empty key with credentials", "/metrics", "Bearer secret", "", 503, "bearer authentication is not configured\n"},
		{"health with empty key", "/healthz", "Bearer ", "", 503, "bearer authentication is not configured\n"},
		{"health without credentials", "/healthz", "", "secret", 400, "bearer header is empty\n"},
		{"health with malformed credentials", "/healthz?probe=1", "invalid", "secret", 400, "token is invalid\n"},
		{"health with wrong credentials", "/healthz", "Bearer wrong", "secret", 403, "forbidden\n"},
		{"health with credentials", "/healthz?probe=1", "Bearer secret", "secret", 202, "accepted"},
		{"health prefix is protected", "/healthz/", "", "secret", 400, "bearer header is empty\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			handler := AuthRequiredWithToken(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				calls++
				w.WriteHeader(http.StatusAccepted)
				_, _ = w.Write([]byte("accepted"))
			}), tt.required)
			request := httptest.NewRequest(http.MethodGet, tt.path, nil)
			request.Header.Set("Authorization", tt.bearer)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			assert.Equal(t, tt.status, response.Code)
			assert.Equal(t, tt.body, response.Body.String())
			assert.Equal(t, tt.status == http.StatusAccepted, calls == 1)
		})
	}
}

func TestRequestLoggingExclusions(t *testing.T) {
	var output bytes.Buffer
	previous := middleware.DefaultLogger
	middleware.DefaultLogger = middleware.RequestLogger(&middleware.DefaultLogFormatter{
		Logger: log.New(&output, "", 0), NoColor: true,
	})
	t.Cleanup(func() { middleware.DefaultLogger = previous })
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Handler", "reached")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("response"))
	})
	logged := RequestLoggerWithoutPaths(next)
	excluded := ExcludePathsFromLogger(map[string]struct{}{"/healthz": {}}, []string{"/quiet/"})(logged)
	for _, tt := range []struct {
		name, path string
		handler    http.Handler
		wantLog    bool
	}{
		{"standalone", "/healthz", logged, true},
		{"ordinary", "/progress", excluded, true},
		{"exact exclusion", "/healthz?probe=1", excluded, false},
		{"not a prefix match", "/healthz/child", excluded, true},
		{"keyword exclusion", "/api/quiet/poll", excluded, false},
		{"keywords ignore query", "/progress?next=/quiet/", excluded, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			output.Reset()
			response := httptest.NewRecorder()
			tt.handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, tt.path, nil))
			assert.Equal(t, http.StatusAccepted, response.Code)
			assert.Equal(t, "reached", response.Header().Get("X-Handler"))
			assert.Equal(t, "response", response.Body.String())
			if tt.wantLog {
				assert.Contains(t, output.String(), tt.path)
				assert.Contains(t, output.String(), "202 8B")
			} else {
				assert.Empty(t, output.String())
			}
		})
	}
}
