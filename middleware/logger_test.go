package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Add Logger test only so that coverage is not unecessarily lowered.
func TestLogger(t *testing.T) {
	assert := assert.New(t)

	helloHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello"))
	})
	handler := Logger(helloHandler)
	respRec := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	assert.NotPanics(func() { handler.ServeHTTP(respRec, req) })
}
