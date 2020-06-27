package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nickhstr/goweb/middleware"
	"github.com/stretchr/testify/assert"
)

// Add Logger test only so that coverage is not unecessarily lowered.
func TestLogger(t *testing.T) {
	assert := assert.New(t)

	helloHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello"))
	})
	opts := middleware.LoggerOptions{
		GitCommit: "abc123",
		Name:      "test",
		Version:   "1.0.0",
	}
	handler := middleware.Logger(opts)(helloHandler)
	respRec := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	assert.NotPanics(func() { handler.ServeHTTP(respRec, req) })
}
