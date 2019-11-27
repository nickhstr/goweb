package middleware_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nickhstr/goweb/middleware"
	"github.com/stretchr/testify/assert"
)

func TestHealth(t *testing.T) {
	helloResp := []byte("Hello world")
	helloHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(helloResp)
	})

	healthCallback := func() map[string]string {
		return map[string]string{
			"health": "all good!",
		}
	}

	expectedHealthResponse := func(v interface{}) []byte {
		resp, _ := json.Marshal(v)

		return resp
	}

	tests := []struct {
		name string
		middleware.HealthConfig
		requestPath  string
		expectedBody []byte
	}{
		{
			"health middleware should respond with the HealthConfig callback",
			middleware.HealthConfig{
				Path:     "/health-test",
				Callback: healthCallback,
			},
			"/health-test",
			expectedHealthResponse(healthCallback()),
		},
		{
			"an empty HealthConfig should return the default response at the default health path",
			middleware.HealthConfig{},
			"/health",
			expectedHealthResponse(map[string]string{}),
		},
		{
			"the wrapped handler should respond to non-health routes",
			middleware.HealthConfig{
				Path:     "/health",
				Callback: healthCallback,
			},
			"/",
			helloResp,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)
			handler := middleware.Health(test.HealthConfig)(helloHandler)
			respRec := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodGet, test.requestPath, nil)
			if err != nil {
				t.Fatal(err)
			}

			handler.ServeHTTP(respRec, req)

			respBody, err := ioutil.ReadAll(respRec.Body)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(test.expectedBody, respBody)
		})
	}
}
