package middleware

import (
	"context"
	"net/http"
	"time"
)

// Timeout adds a timeout to each request's context.
// If the duration is zero, Timeout defaults to
// fifteen seconds.
func Timeout(d time.Duration) Middleware {
	timeout := d

	if d == 0 {
		timeout = 15 * time.Second
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
