package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/nickhstr/goweb/logger"
)

// LoggerOptions holds all logger middleware options.
type LoggerOptions struct {
	GitCommit string
	Name      string
	Version   string
}

// Logger outputs general information about requests.
func Logger(opts LoggerOptions) Middleware {
	return func(next http.Handler) http.Handler {
		var log = logger.New("middleware")

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, 0)
			start := time.Now()

			// add logger to request's context
			l := log.With().
				Str("method", r.Method).
				Str("url", r.URL.String()).
				Str("host", r.Host).
				Interface("requestHeaders", r.Header).
				Str("appName", opts.Name).
				Str("gitRevision", opts.GitCommit).
				Logger()
			r = r.WithContext(l.WithContext(r.Context()))

			next.ServeHTTP(ww, r)

			status := ww.Status()
			responseTime := time.Since(start).String()

			if status >= http.StatusInternalServerError {
				l.Error().Int("status", status).Str("responseTime", responseTime).Msg("Route handler")
			} else {
				l.Info().Int("status", status).Str("responseTime", responseTime).Msg("Route handler")
			}
		})
	}
}
