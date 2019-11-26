package middleware

import (
	"net/http"
	"reflect"

	"github.com/nickhstr/goweb/env"
	"github.com/unrolled/secure"
)

// Secure creates the security middleware
func Secure(options secure.Options) Middleware {
	return func(handler http.Handler) http.Handler {
		// The performance impact of using reflect here is negligible, as it's only
		// used once when getting the middleware handler
		if reflect.DeepEqual(options, secure.Options{}) {
			options = secure.Options{
				BrowserXssFilter:        true,
				ContentTypeNosniff:      true,
				CustomFrameOptionsValue: "SAMEORIGIN",
				STSSeconds:              180 * 24 * 60 * 60, // Default to 180 days
				STSIncludeSubdomains:    true,
			}
		}

		if !env.IsProd() {
			options.IsDevelopment = true
		}

		return secure.New(options).Handler(handler)
	}
}
