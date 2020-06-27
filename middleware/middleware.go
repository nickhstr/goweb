// Package middleware provides multiple middlewares, useful for any HTTP service.
// These middlewares are not specific to any application, and are made to be as
// reusable and idiomatic as possible.
// Compatibility is guaranteed for `net/http`, however these middlewares should
// be compatible with any third-party package which conforms to the standard
// library's APIs.
package middleware

import (
	"net/http"
)

// Middleware aliases the function type for all middleware.
type Middleware = func(http.Handler) http.Handler

// Compose adds middleware handlers to a given handler.
// Middleware handlers are ordered from first to last.
func Compose(h http.Handler, m ...Middleware) http.Handler {
	// Execute middleware by LIFO order, so that consumers of Compose can order
	// their m from first to last
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}

	return h
}
