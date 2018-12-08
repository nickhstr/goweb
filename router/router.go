package router

import (
	"net/http"

	"github.com/go-chi/chi"
)

// New returns a new router.
// The return type has been explicitly set as an http.Handler to enforce Route
// registration via this New function. By doing so, all route handlers are
// guaranteed to be of the same http.HandlerFunc type, rather than the
// httprouter.Handle type.
// Using the standard library's types and interfaces is desirable as it allows
// more flexibility in which router to use, keeps handlers standardized with
// the more widely understood net/http package, and ensures access to httprouter's
// Params is only available via an *http.Request's context.
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
