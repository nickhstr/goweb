package middleware

import (
	"net/http"
	"time"

	"github.com/nickhstr/goweb/env"
	"github.com/nickhstr/goweb/logger"
	"github.com/rs/zerolog/hlog"
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

// Flush satisfies the http.Flusher interface
func (sr *statusRecorder) Flush() {
	if w, ok := sr.ResponseWriter.(http.Flusher); ok {
		w.Flush()
	}
}

// Logger outputs general information about requests.
func Logger(handler http.Handler) http.Handler {
	var log = logger.New("middleware")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use 200 status code as default
		rw := &statusRecorder{w, http.StatusOK}
		start := time.Now()

		// add logger to request's context
		l := log.With().Logger()
		r = r.WithContext(l.WithContext(r.Context()))

		handler.ServeHTTP(rw, r)

		hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Str("url", r.URL.String()).
			Str("host", r.Host).
			Interface("request-headers", r.Header).
			Str("response-time", time.Since(start).String()).
			Int("status", rw.status).
			Str("app-name", env.Get("APP_NAME", "web-service")).
			Msg("Route handler")
	})
}
