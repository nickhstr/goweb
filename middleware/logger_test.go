package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Testing the log output doesn't add much value; however, testing the statusRecorder
// used by the Logging middleware is useful.
func TestStatusRecorder(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		msg string
		http.Handler
		expectedCode int
	}{
		{
			"the statusRecorder should record status code set explicitly",
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte("Not found"))
			}),
			http.StatusNotFound,
		},
	}

	for _, test := range tests {
		respRec := httptest.NewRecorder()
		sr := &statusRecorder{ResponseWriter: respRec}

		req, err := http.NewRequest(http.MethodGet, "/", nil)
		if err != nil {
			t.Fatal(err)
		}

		test.Handler.ServeHTTP(sr, req)

		assert.Equal(test.expectedCode, sr.status, test.msg)
	}
}

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
