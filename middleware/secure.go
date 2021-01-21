package middleware

import (
	"reflect"

	"github.com/nickhstr/goweb/config"
	"github.com/unrolled/secure"
)

var secureDefaultOpts = secure.Options{
	BrowserXssFilter:        true,
	ContentTypeNosniff:      true,
	CustomFrameOptionsValue: "SAMEORIGIN",
	STSSeconds:              180 * 24 * 60 * 60, // Default to 180 days
	STSIncludeSubdomains:    true,
}

// Secure creates the security middleware
func Secure(opts secure.Options) Middleware {
	// The performance impact of using reflect here is negligible, as it's only
	// used once when getting the middleware handler
	if reflect.DeepEqual(opts, secure.Options{}) {
		opts = secureDefaultOpts
	}

	if !config.IsProd() {
		opts.IsDevelopment = true
	}

	return secure.New(opts).Handler
}

// SecureDefault is similar to Secure, but provides sane security defaults,
// without requiring options to be provided.
func SecureDefault() Middleware {
	opts := secureDefaultOpts
	if !config.IsProd() {
		opts.IsDevelopment = true
	}

	return secure.New(opts).Handler
}
