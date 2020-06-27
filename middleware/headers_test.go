package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nickhstr/goweb/middleware"
	"github.com/stretchr/testify/assert"
)

func TestAppHeaders(t *testing.T) {
	helloResp := []byte("Hello world")
	helloHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(helloResp)
	})

	tests := []struct {
		name string
		middleware.AppHeadersOptions
		expectedHeaders map[string]string
	}{
		{
			"the headers middleware should add headers from the config",
			middleware.AppHeadersOptions{
				Name:      "test-app",
				GitCommit: "abc123",
				Region:    "local",
				Version:   "1.0.0",
			},
			map[string]string{
				http.CanonicalHeaderKey("app-name"):    "test-app",
				http.CanonicalHeaderKey("app-version"): "1.0.0-abc123",
				http.CanonicalHeaderKey("region"):      "local",
			},
		},
		{
			"the headers middleware should use GitCommit for its version",
			middleware.AppHeadersOptions{
				Name:      "test-app",
				GitCommit: "abc123",
				Region:    "local",
				Version:   "abc123efg456",
			},
			map[string]string{
				http.CanonicalHeaderKey("app-name"):    "test-app",
				http.CanonicalHeaderKey("app-version"): "abc123",
				http.CanonicalHeaderKey("region"):      "local",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)
			matchingHeaders := true

			handler := middleware.AppHeaders(test.AppHeadersOptions)(helloHandler)
			respRec := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodGet, "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			handler.ServeHTTP(respRec, req)

			for key, val := range respRec.Header() {
				if val[0] != test.expectedHeaders[key] {
					matchingHeaders = false
				}
			}

			assert.True(matchingHeaders)
		})
	}
}

func TestSetHeaders(t *testing.T) {
	helloResp := []byte("Hello world")
	helloHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(helloResp)
	})

	tests := []struct {
		name  string
		key   string
		value string
	}{
		{
			"Response headers should match supplied headers",
			"Content-Type",
			"application/json",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)
			handler := middleware.SetHeader(test.key, test.value)(helloHandler)
			respRec := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodGet, "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			handler.ServeHTTP(respRec, req)
			result := respRec.Result()

			assert.Equal(result.Header.Get(test.key), test.value)
		})
	}
}
