package middleware

import (
	"net/http"

	"github.com/nickhstr/goweb/env"
	"github.com/unrolled/secure"
)

// Secure creates the security middleware
func Secure(options secure.Options) Middleware {
	return func(handler http.Handler) http.Handler {
		if !env.Prod() {
			options.IsDevelopment = true
		}

		secureMiddleware := secure.New(options)

		return secureMiddleware.Handler(handler)
	}
}
