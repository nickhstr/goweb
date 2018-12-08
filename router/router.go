package router

import (
	"net/http"

	"github.com/go-chi/chi"
)

// New creates a new router, fully compatible with net/http.
// Maintaining this compatibility allows flexibility in choosing a routing library,
// without sacrificing the ability to use net/http-compatible packages.
func New(routes []Route) http.Handler {
	router := chi.NewRouter()

	// Register routes with router
	for _, route := range routes {
		var handler http.Handler
		handler = route.Handler

		router.Method(route.Method, route.Path, handler)
	}

	return router
}

// Route defines the fundamental pieces of information
// required of every route.
type Route struct {
	Method  string
	Path    string
	Handler http.HandlerFunc
}
