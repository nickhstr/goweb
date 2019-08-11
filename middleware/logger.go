package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/nickhstr/goweb/env"
)

// Used to wrap an http.ResponseWriter to capture the response's status code
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}

// Logger outputs general information about requests.
func Logger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use 200 status code as default
		rw := &statusRecorder{w, http.StatusOK}
		start := time.Now()

		handler.ServeHTTP(rw, r)

		log.Info().
			Str("method", r.Method).
			Str("url", r.URL.Path).
			Str("host", r.Host).
			Interface("request-headers", r.Header).
			Str("response-time", fmt.Sprintf("%v", time.Since(start))).
			Int("status", rw.status).
			Str("app-name", env.Get("APP_NAME", "web-service")).
			Msg("Route handler")
	})
}
