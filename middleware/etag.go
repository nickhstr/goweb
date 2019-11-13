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

// Flush satisfies the http.Flusher interface
func (e *EtagWriter) Flush() {
	if w, ok := e.ResponseWriter.(http.Flusher); ok {
		w.Flush()
	}
}

func (e *EtagWriter) Write(p []byte) (int, error) {
	// Don't generate etag for error status codes
	if e.status >= 400 {
		e.ResponseWriter.WriteHeader(e.status)
		return e.ResponseWriter.Write(p)
	}

	// `weak` should be true when the ResponseWriter is a Flusher, as that
	// indicates the response body can be streamed
	_, weak := e.ResponseWriter.(http.Flusher)
	et := etag.Generate(p, weak)
	e.ResponseWriter.Header().Set("ETag", et)

	if e.clientEtag == et {
		// Set status to not modified, and write empty body
		e.ResponseWriter.WriteHeader(http.StatusNotModified)
		return e.ResponseWriter.Write([]byte{})
	}

	if e.status != 0 {
		// Make sure to call ResponseWriter's WriteHeader, with stored status
		e.ResponseWriter.WriteHeader(e.status)
	}

	return e.ResponseWriter.Write(p)
}

// Etag middleware sets the ETag header.
func Etag(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e := &EtagWriter{
			ResponseWriter: w,
			clientEtag:     r.Header.Get("If-None-Match"),
		}

		handler.ServeHTTP(e, r)
	})
}
