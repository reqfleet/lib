// Derived from reqfleet/shibuya; modified for lib.
package auth

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAuthRequiredWithToken_DoesNotLeakProvidedToken(t *testing.T) {
	handler := AuthRequiredWithToken(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), "expected-token")

	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer user-provided-secret")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, recorder.Code)
	}

	body, err := io.ReadAll(recorder.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	bodyString := string(body)
	if strings.Contains(bodyString, "user-provided-secret") {
		t.Fatalf("response leaked token: %s", bodyString)
	}
	if !strings.Contains(strings.ToLower(bodyString), "forbidden") {
		t.Fatalf("expected generic forbidden response, got: %s", bodyString)
	}
}
