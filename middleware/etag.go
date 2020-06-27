package middleware

import (
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/nickhstr/goweb/etag"
)

// EtagWriter provides a ResponseWriter which holds info needed for verifying etags.
type EtagWriter struct {
	http.ResponseWriter
	clientEtag string
}

// NewEtagWriter returns an EtagWriter with a wrapped ResponseWriter.
func NewEtagWriter(w http.ResponseWriter) *EtagWriter {
	return &EtagWriter{
		ResponseWriter: middleware.NewWrapResponseWriter(w, 0),
	}
}

// ClientEtag sets the requested ETag.
func (e *EtagWriter) ClientEtag(t string) *EtagWriter {
	e.clientEtag = t

	return e
}

// Write overrides the wrapped http.ResponseWriter's Write method, to
// provide etag-specific headers.
func (e *EtagWriter) Write(p []byte) (int, error) {
	rw, ok := e.ResponseWriter.(middleware.WrapResponseWriter)
	if !ok {
		return e.Write(p)
	}

	// Don't generate etag for error status codes
	if rw.Status() >= http.StatusBadRequest {
		return rw.Write(p)
	}

	// `weak` should be true when the ResponseWriter is a Flusher, as that
	// indicates the response body can be streamed
	_, weak := rw.(http.Flusher)
	et := etag.Generate(p, weak)
	rw.Header().Set("ETag", et)

	if e.clientEtag == et {
		// Set status to not modified, and return
		rw.WriteHeader(http.StatusNotModified)
		return 0, nil
	}

	return rw.Write(p)
}

// Etag middleware sets the ETag header.
func Etag(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e := NewEtagWriter(w).ClientEtag(r.Header.Get("If-None-Match"))

		next.ServeHTTP(e, r)
	})
}
