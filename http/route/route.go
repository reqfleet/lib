// Derived from reqfleet/shibuya; modified for lib.
package route

import (
	"net/http"
	"strings"
)

type Route struct {
	Name        string
	Method      string
	Path        string
	HandlerFunc http.HandlerFunc
}

type Routes []*Route

// Router assembles routes before serving. AddRoutes and Mount mutate route
// paths; finish registration before calling Mux and do not reuse mounted routes.
type Router struct {
	Name      string
	Path      string
	routes    Routes
	subrouter []*Router
}

func NewRouter(name, path string) *Router {
	return &Router{Name: name, Path: path}
}

func (r *Route) MakePattern() string {
	return strings.ToUpper(r.Method) + " " + r.Path
}

func (rou *Router) MakeFullPath(path string) string {
	path = strings.TrimPrefix(path, "/")
	return strings.TrimSuffix(rou.Path+"/"+path, "/")
}

func (rou *Router) AddRoutes(routes Routes) {
	for _, r := range routes {
		r.Path = rou.MakeFullPath(r.Path)
		rou.routes = append(rou.routes, r)
	}
}

func (rou *Router) Mount(sub *Router) {
	for _, r := range sub.GetRoutes() {
		r.Path = rou.MakeFullPath(r.Path)
	}
	rou.subrouter = append(rou.subrouter, sub)
}

func (rou *Router) GetRoutes() Routes {
	all := make(Routes, 0)
	all = append(all, rou.routes...)
	for _, r := range rou.subrouter {
		all = append(all, r.GetRoutes()...)
	}
	return all
}

func (rou *Router) Mux() *http.ServeMux {
	mux := http.NewServeMux()
	for _, r := range rou.GetRoutes() {
		mux.HandleFunc(r.MakePattern(), r.HandlerFunc)
	}
	return mux
}
