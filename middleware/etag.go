package middleware

import (
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/nickhstr/goweb/etag"
)

// EtagWriter provides a ResponseWriter which holds info needed for verifying etags.
type EtagWriter struct {
	middleware.WrapResponseWriter
	clientEtag string
}

func (e *EtagWriter) Write(p []byte) (int, error) {
	// Don't generate etag for error status codes
	if e.Status() >= 400 {
		e.WrapResponseWriter.WriteHeader(e.Status())
		return e.WrapResponseWriter.Write(p)
	}

	// `weak` should be true when the ResponseWriter is a Flusher, as that
	// indicates the response body can be streamed
	_, weak := e.WrapResponseWriter.(http.Flusher)
	et := etag.Generate(p, weak)
	e.WrapResponseWriter.Header().Set("ETag", et)

	if e.clientEtag == et {
		// Set status to not modified, and write empty body
		e.WrapResponseWriter.WriteHeader(http.StatusNotModified)
		return e.WrapResponseWriter.Write([]byte{})
	}

	if e.Status() != 0 {
		// Make sure to call ResponseWriter's WriteHeader, with stored status
		e.WrapResponseWriter.WriteHeader(e.Status())
	}

	return e.WrapResponseWriter.Write(p)
}

// Etag middleware sets the ETag header.
func Etag(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e := &EtagWriter{
			WrapResponseWriter: middleware.NewWrapResponseWriter(w, 0),
			clientEtag:         r.Header.Get("If-None-Match"),
		}

		handler.ServeHTTP(e, r)
	})
}
