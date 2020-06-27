package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/nickhstr/goweb/env"
	"github.com/nickhstr/goweb/write"
	"github.com/rs/zerolog/hlog"
)

// Recover middleware recovers from panics, and logs the error.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				errMsg := "An unexpected error occured"
				stack := debug.Stack()
				hlog.FromRequest(r).Error().
					Bytes("stacktrace", stack).
					Msg(errMsg)

				if env.IsProd() {
					write.Error(w, errMsg, http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "text/html; charset=UTF-8")
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "<h1>panic: %v</h1><pre>%s</pre>", err, stack)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
