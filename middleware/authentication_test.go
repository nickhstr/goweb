package middleware_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nickhstr/goweb/middleware"
	"github.com/nickhstr/goweb/write"
	"github.com/stretchr/testify/assert"
)

func TestAuth(t *testing.T) {
	helloResp := []byte("Hello world")
	helloHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(helloResp)
	})

	errResponse := func(err string) []byte {
		resp, _ := json.Marshal(write.ErrorResponse{
			Status:     http.StatusUnauthorized,
			StatusText: http.StatusText(http.StatusUnauthorized),
			Error:      err,
		})
		// Add newline, as write.Error uses json.Encoder, which adds a newline
		resp = append(resp, '\n')

		return resp
	}

	tests := []struct {
		name string
		middleware.AuthOptions
		requestPath  string
		expectedBody []byte
	}{
		{
			"a request with secret key should return handler's response",
			middleware.AuthOptions{
				SecretKey:    "supersecret",
				ErrorMessage: "Ah ah ah ah ahh, you didn't say the magic word",
			},
			"/?apiKey=supersecret",
			helloResp,
		},
		{
			"a request without secret key should return error response",
			middleware.AuthOptions{
				SecretKey:    "supersecret",
				ErrorMessage: "Ah ah ah ah ahh, you didn't say the magic word",
			},
			"/?apiKey=blah",
			errResponse("Ah ah ah ah ahh, you didn't say the magic word"),
		},
		{
			"default error message should be used when one is not supplied to AuthOptions",
			middleware.AuthOptions{
				SecretKey: "supersecret",
			},
			"/?apiKey=blah",
			errResponse("invalid API key supplied"),
		},
		{
			"supplied handler's response should be served when no secret key is set",
			middleware.AuthOptions{},
			"/",
			helloResp,
		},
		{
			"whitelisted routes should not require authentication",
			middleware.AuthOptions{
				SecretKey: "supersecret",
				WhiteList: []string{"/hello"},
			},
			"/hello",
			helloResp,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)
			respRec := httptest.NewRecorder()
			handler := middleware.Auth(test.AuthOptions)(helloHandler)

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
