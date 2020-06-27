package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/gorilla/mux"
	"github.com/newrelic/go-agent/v3/newrelic"
	nr "github.com/nickhstr/goweb/newrelic"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// startTime is used to mark when the consuming application has started.
var startTime time.Time

func init() {
	startTime = time.Now()
}

// Route offers a flexible way to add route handling to a given router.
type Route func(*mux.Router)

// DefaultRoutesOptions provides configuration for the default routes.
type DefaultRoutesOptions struct {
	HealthOptions
	PrometheusOptions
}

// DefaultRoutes returns a slice of opinionated routes which most
// wxu-web services use.
func DefaultRoutes(opts DefaultRoutesOptions) []Route {
	return []Route{
		GetPrometheusRoute(opts.PrometheusOptions),
		GetHealthRoute(opts.HealthOptions),
		DebugRoute,
		NotFound,
		MethodNotAllowed,
	}
}

// PrometheusOptions exposes options to supply the Prometheus route.
type PrometheusOptions struct {
	// Path is the URL path for the prometheus metrics route.
	Path string
}

// GetPrometheusRoute returns the Prometheus metrics route.
func GetPrometheusRoute(opts PrometheusOptions) Route {
	return func(r *mux.Router) {
		r.Handle(opts.Path, promhttp.Handler()).Methods(http.MethodGet)
	}
}

// HealthOptions holds all values needed for the Health route.
type HealthOptions struct {
	// GitCommit is the application's current git commit ID.
	GitCommit string

	// Name is the name of the application.
	Name string

	// Path is the URL path for the health check route.
	Path string

	// Region is the server region where the app is currently running.
	Region string

	// Version is the version of the application. There is currently no
	// standard for this; convention has been semver.
	Version string
}

// GetHealthRoute returns the health check route.
func GetHealthRoute(opts HealthOptions) Route {
	type response struct {
		Sha1    string `json:"sha1"`
		Name    string `json:"name"`
		Region  string `json:"region"`
		Uptime  string `json:"uptime"`
		Version string `json:"version"`
	}

	return func(r *mux.Router) {
		r.HandleFunc(opts.Path, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			v := response{
				Sha1:    opts.GitCommit,
				Name:    opts.Name,
				Region:  opts.Region,
				Uptime:  fmt.Sprintf("%fs", time.Since(startTime).Seconds()),
				Version: opts.Version,
			}
			enc := json.NewEncoder(w)
			_ = enc.Encode(v)
		}).Methods(http.MethodGet)
	}
}

// DebugRoute sets up the /debug/pprof-related routes.
func DebugRoute(r *mux.Router) {
	dr := r.PathPrefix("/debug/pprof").Subrouter().StrictSlash(true)
	dr.HandleFunc("/", pprof.Index)
	dr.HandleFunc("/allocs", pprof.Index)
	dr.HandleFunc("/block", pprof.Index)
	dr.HandleFunc("/cmdline", pprof.Cmdline)
	dr.HandleFunc("/goroutine", pprof.Index)
	dr.HandleFunc("/heap", pprof.Index)
	dr.HandleFunc("/mutex", pprof.Index)
	dr.HandleFunc("/profile", pprof.Profile)
	dr.HandleFunc("/symbol", pprof.Symbol)
	dr.HandleFunc("/threadcreate", pprof.Index)
	dr.HandleFunc("/trace", pprof.Trace)
}

// NotFound adds a NotFoundHandler to the router, instrumented by New Relic.
func NotFound(r *mux.Router) {
	_, handler := newrelic.WrapHandle(nr.App(), "NotFoundHandler", http.NotFoundHandler())
	r.NotFoundHandler = handler
}

// MethodNotAllowed adds a MethodNotAllowedHandler to the router, instrumented by New Relic.
func MethodNotAllowed(r *mux.Router) {
	_, handler := newrelic.WrapHandle(nr.App(), "MethodNotAllowedHandler", http.HandlerFunc(methodNotAllowed))
	r.MethodNotAllowedHandler = handler
}

// methodNotAllowed replies to the request with an HTTP status code 405.
func methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}
