package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/nickhstr/goweb/env"
)

// Recover middleware recovers from panics, and logs the error.
func Recover(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				errMsg := "An unexpected error occured"
				stack := debug.Stack()
				log.Error().
					Bytes("stacktrace", stack).
					Msg(errMsg)

				if env.Prod() {
					http.Error(w, errMsg, http.StatusInternalServerError)
					return
				}

				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "<h1>panic: %v</h1><pre>%s</pre>", err, string(stack))
			}
		}()

		handler.ServeHTTP(w, r)
	})
}
