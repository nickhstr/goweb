package middleware

import (
	"net/http"

	"github.com/nickhstr/goweb/etag"
)

type etagWriter struct {
	http.ResponseWriter
	status int
}

// WriteHeader must be defined to override the ResponseWriter's WriteHeader.
// This is needed to set the etag header in etagWriter.Write, as ResponseWriter's
// WriteHeader prevents setting headers after it is called.
func (e *etagWriter) WriteHeader(status int) {
	e.status = status
}

func (e *etagWriter) Write(p []byte) (int, error) {
	et := etag.Generate(p, false)
	e.ResponseWriter.Header().Set("ETag", et)
	// Make sure to call ResponseWriter's WriteHeader, with stored status
	e.ResponseWriter.WriteHeader(e.status)

	return e.ResponseWriter.Write(p)
}

// Etag middleware sets the ETag header. Note, this middleware is only useful for buffered response
// bodies.
func Etag(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e := &etagWriter{ResponseWriter: w}

		handler.ServeHTTP(e, r)
	})
}
