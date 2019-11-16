package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/nickhstr/goweb/env"
	"github.com/nickhstr/goweb/logger"
	"github.com/rs/zerolog/hlog"
)

// Logger outputs general information about requests.
func Logger(handler http.Handler) http.Handler {
	var log = logger.New("middleware")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, 0)
		start := time.Now()

		// add logger to request's context
		l := log.With().Logger()
		r = r.WithContext(l.WithContext(r.Context()))

		handler.ServeHTTP(ww, r)

		hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Str("url", r.URL.String()).
			Str("host", r.Host).
			Interface("request-headers", r.Header).
			Str("response-time", time.Since(start).String()).
			Int("status", ww.Status()).
			Str("app-name", env.Get("APP_NAME", "web-service")).
			Msg("Route handler")
	})
}
