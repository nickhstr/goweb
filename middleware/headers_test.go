package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nickhstr/goweb/middleware"
	"github.com/stretchr/testify/assert"
)

func TestHeaders(t *testing.T) {
	assert := assert.New(t)

	helloResp := []byte("Hello world")
	helloHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(helloResp)
	})

	tests := []struct {
		msg string
		middleware.AppHeaders
		expectedHeaders map[string]string
	}{
		{
			"the headers middleware should add headers from the config",
			middleware.AppHeaders{
				AppName:     "test-app",
				AppVersion:  "1.0.0",
				GitRevision: "abc123",
				Region:      "local",
			},
			map[string]string{
				http.CanonicalHeaderKey("app-name"):    "test-app",
				http.CanonicalHeaderKey("app-version"): "1.0.0-abc123",
				http.CanonicalHeaderKey("region"):      "local",
			},
		},
	}

	for _, test := range tests {
		var matchingHeaders = true

		handler := middleware.Headers(test.AppHeaders)(helloHandler)
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

		assert.True(matchingHeaders, test.msg)
	}
}
