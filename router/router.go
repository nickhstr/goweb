// Package router provides many router utilities, with the primary goals of
// ease of use, flexibility, and http.Handler interface compliance.
package router

import (
	"net/http"
	"path"

	"github.com/gorilla/mux"
	"github.com/nickhstr/goweb/middleware"
)

// DefaultOptions provides a limited set of options for
// the Default router.
type DefaultOptions struct {
	// Auth can be set to true to enable query-param-based authentication.
	Auth bool

	// Compress can be set to true to enable compression for all responses.
	Compress bool

	// CORS can be set to true to enable CORS middleware.
	CORS bool

	// ETag, if set to true, will handle etag-related headers accordingly.
	ETag bool

	// GitCommit is the git SHA of the app's current git commit.
	GitCommit string

	// Name is the application name.
	Name string

	// Region is the cluster region where the app is running.
	Region string

	// Version is the current version of the application.
	Version string

	// Whitelist is a string of regular expression strings,
	// matching URL paths which should not require authentication.
	WhiteList []string
}

// Default is the same as New, but adds some default router
// middleware and routes most services use.
func Default(routes []Route, opts DefaultOptions) *mux.Router {
	healthPath := path.Join("/", opts.Name, "health")

	ro := DefaultRoutesOptions{
		HealthOptions: HealthOptions{
			GitCommit: opts.GitCommit,
			Name:      opts.Name,
			Path:      healthPath,
			Region:    opts.Region,
			Version:   opts.Version,
		},
		PrometheusOptions: PrometheusOptions{
			Path: "/metrics",
		},
	}
	mwo := DefaultMiddlewareOptions{
		AuthOptions: AuthOptions{
			Enabled: opts.Auth,
			WhiteList: append(
				opts.WhiteList,
				`^`+healthPath+`$`,
				`^/debug/pprof.*`,
			),
		},
		Compress:  opts.Compress,
		CORS:      opts.CORS,
		ETag:      opts.ETag,
		GitCommit: opts.GitCommit,
		Name:      opts.Name,
		Region:    opts.Region,
		Version:   opts.Version,
	}

	rs := append(DefaultRoutes(ro), routes...)
	r := New(rs)

	for _, mw := range DefaultMiddleware(mwo) {
		// Rather than unpack middleware, we iterate this way as
		// the []middleware.Middleware type does not match the
		// []mux.MiddlewareFunc type, and slices do not allow
		// type conversion
		r.Use(mw)
	}

	return r
}

// DefaultMux creates and returns a Default router, with some
// some application-wide middleware.
func DefaultMux(routes []Route, opts DefaultOptions) http.Handler {
	r := Default(routes, opts)
	h := middleware.Compose(
		r,
		middleware.Logger(middleware.LoggerOptions{
			GitCommit: opts.GitCommit,
			Name:      opts.Name,
			Version:   opts.Version,
		}),
		middleware.Recover,
	)

	return h
}

// New creates a new router, fully compatible with http.Handler.
// Maintaining this compatibility allows flexibility in choosing a routing
// library, without sacrificing the ability to use net/http-compatible packages.
func New(routes []Route) *mux.Router {
	r := mux.NewRouter()
	RegisterRoutes(r, routes...)

	return r
}

// RegisterRoutes provides an easy-to-use way of registering many
// Routes at once.
func RegisterRoutes(r *mux.Router, routes ...Route) {
	for _, route := range routes {
		route(r)
	}
}
