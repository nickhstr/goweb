package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nickhstr/goweb/middleware"
	"github.com/stretchr/testify/assert"
)

func TestRecover(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		msg string
		http.Handler
		expectedStatus int
	}{
		{
			"a handler which panics should result in an error status code",
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic("ruh roh")
			}),
			http.StatusInternalServerError,
		},
		{
			"response from successful handler should not be modified",
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("all good"))
			}),
			http.StatusOK,
		},
	}

	for _, test := range tests {
		handler := middleware.Recover(test.Handler)
		respRec := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "http://foo.com", nil)
		if err != nil {
			t.Fatal(err)
		}

		assert.NotPanics(func() { handler.ServeHTTP(respRec, req) }, test.msg)
		assert.Equal(test.expectedStatus, respRec.Result().StatusCode)
	}
}
