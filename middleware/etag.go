package middleware

import (
	"net/http"

	"github.com/nickhstr/goweb/etag"
)

// EtagWriter provides a ResponseWriter which holds info needed for verifying etags.
type EtagWriter struct {
	http.ResponseWriter
	status     int
	clientEtag string
}

// WriteHeader must be defined to override the ResponseWriter's WriteHeader.
// This is needed to set the etag header in etagWriter.Write, as ResponseWriter's
// WriteHeader prevents setting headers after it is called.
func (e *EtagWriter) WriteHeader(status int) {
	e.status = status
}

func (e *EtagWriter) Write(p []byte) (int, error) {
	et := etag.Generate(p, false)
	e.ResponseWriter.Header().Set("ETag", et)

	if e.clientEtag == et {
		// Set status to not modified, and write empty body
		e.ResponseWriter.WriteHeader(http.StatusNotModified)
		return e.ResponseWriter.Write([]byte{})
	}

	// Make sure to call ResponseWriter's WriteHeader, with stored status
	e.ResponseWriter.WriteHeader(e.status)
	return e.ResponseWriter.Write(p)
}

// Etag middleware sets the ETag header. Note, this middleware is only useful for buffered response
// bodies.
func Etag(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e := &EtagWriter{
			ResponseWriter: w,
			clientEtag:     r.Header.Get("If-None-Match"),
		}

		handler.ServeHTTP(e, r)
	})
}
