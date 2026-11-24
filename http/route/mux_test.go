package route_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/reqfleet/lib/http/route"
	"github.com/stretchr/testify/assert"
)

func TestMountedMuxDispatch(t *testing.T) {
	root := route.NewRouter("root", "")
	api := route.NewRouter("api", "/api")
	users := route.NewRouter("users", "/users")
	users.AddRoutes(route.Routes{{
		Name: "user", Method: "get", Path: "/{user}",
		HandlerFunc: func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-User", r.PathValue("user"))
			w.WriteHeader(http.StatusAccepted)
		},
	}})
	api.Mount(users)
	root.Mount(api)
	mux := root.Mux()
	for _, tt := range []struct {
		method, path string
		status       int
	}{
		{http.MethodGet, "/api/users/andy", http.StatusAccepted},
		{http.MethodHead, "/api/users/andy", http.StatusAccepted},
		{http.MethodPost, "/api/users/andy", http.StatusMethodNotAllowed},
		{http.MethodGet, "/api/users/andy/extra", http.StatusNotFound},
	} {
		response := httptest.NewRecorder()
		mux.ServeHTTP(response, httptest.NewRequest(tt.method, tt.path, nil))
		assert.Equal(t, tt.status, response.Code)
		if tt.status == http.StatusAccepted {
			assert.Equal(t, "andy", response.Header().Get("X-User"))
		}
	}
}
