// Derived from reqfleet/shibuya; modified for lib.
package auth

import (
	"context"
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
	authtoken "github.com/reqfleet/lib/http/auth/token"
)

type contextKey string

const isExcludedKey contextKey = "isExcluded"

// AuthRequiredWithToken requires an exact Bearer token on every request.
// Missing or malformed credentials return 400; mismatched tokens return 403.
// An empty requiredToken disables access: every request returns 401 without
// invoking next. Applications must route public endpoints outside this wrapper.
func AuthRequiredWithToken(next http.Handler, requiredToken string) http.Handler {
	if requiredToken == "" {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "bearer authentication is not configured", http.StatusUnauthorized)
		})
	}
	fn := func(w http.ResponseWriter, r *http.Request) {
		token, err := authtoken.FindBearerToken(r.Header)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if subtle.ConstantTimeCompare([]byte(token), []byte(requiredToken)) != 1 {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// RequestLoggerWithoutPaths logs requests unless ExcludePathsFromLogger marked
// them as excluded. Without that middleware, every request is logged.
func RequestLoggerWithoutPaths(next http.Handler) http.Handler {
	logged := middleware.Logger(next)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isExcluded, _ := r.Context().Value(isExcludedKey).(bool)
		if isExcluded {
			next.ServeHTTP(w, r)
			return
		}
		logged.ServeHTTP(w, r)
	})
}

// ExcludePathsFromLogger must wrap RequestLoggerWithoutPaths. Paths match
// exactly; keywords match any substring of URL.Path. Do not mutate the supplied
// map or slice after constructing the middleware.
func ExcludePathsFromLogger(excludedPaths map[string]struct{}, excludedKeywords []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			isExcluded := false
			if _, ok := excludedPaths[r.URL.Path]; ok {
				isExcluded = true
			}
			if !isExcluded {
				for _, k := range excludedKeywords {
					if strings.Contains(r.URL.Path, k) {
						isExcluded = true
						break
					}
				}
			}
			contextR := r.WithContext(context.WithValue(r.Context(), isExcludedKey, isExcluded))
			next.ServeHTTP(w, contextR)
		}
		return http.HandlerFunc(fn)
	}
}
