package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/nickhstr/goweb/etag"
	"github.com/nickhstr/goweb/middleware"
	"github.com/stretchr/testify/assert"
)

func TestEtag(t *testing.T) {
	tests := []struct {
		name string
		*http.Request
		handler        http.Handler
		expectedStatus int
		expectedEtag   string
	}{
		{
			"200 response should include an etag",
			&http.Request{
				Method: http.MethodGet,
				URL: &url.URL{
					Scheme: "http",
					Host:   "foo.com",
					Path:   "/good",
				},
				Header: http.Header{
					"If-None-Match": []string{""},
				},
			},
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("all good here"))
			}),
			http.StatusOK,
			etag.Generate([]byte("all good here"), true),
		},
		{
			"400 response should not include an etag",
			&http.Request{
				Method: http.MethodGet,
				URL: &url.URL{
					Scheme: "http",
					Host:   "foo.com",
					Path:   "/notfound",
				},
				Header: http.Header{
					"If-None-Match": []string{""},
				},
			},
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte("not found"))
			}),
			http.StatusNotFound,
			"",
		},
		{
			"500 response should not include an etag",
			&http.Request{
				Method: http.MethodGet,
				URL: &url.URL{
					Scheme: "http",
					Host:   "foo.com",
					Path:   "/error",
				},
				Header: http.Header{
					"If-None-Match": []string{""},
				},
			},
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("server error"))
			}),
			http.StatusInternalServerError,
			"",
		},
		{
			"200 response with matching if-none-match etag should return not modified status code",
			&http.Request{
				Method: http.MethodGet,
				URL: &url.URL{
					Scheme: "http",
					Host:   "foo.com",
					Path:   "/not-modified",
				},
				Header: http.Header{
					"If-None-Match": []string{etag.Generate([]byte("content not modified"), true)},
				},
			},
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("content not modified"))
			}),
			http.StatusNotModified,
			etag.Generate([]byte("content not modified"), true),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			handler := middleware.Etag(test.handler)
			respRec := httptest.NewRecorder()

			handler.ServeHTTP(respRec, test.Request)
			resp := respRec.Result()
			e := resp.Header.Get("etag")

			assert.Equal(test.expectedEtag, e)
			assert.Equal(test.expectedStatus, resp.StatusCode)
		})
	}
}
