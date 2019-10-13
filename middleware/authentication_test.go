package middleware

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuth(t *testing.T) {
	assert := assert.New(t)

	helloResp := []byte("Hello world")
	helloHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(helloResp)
	})

	errResponse := func(msg string) []byte {
		resp, _ := json.Marshal(unauthError{msg})

		return resp
	}

	tests := []struct {
		msg string
		AuthConfig
		requestPath  string
		shouldError  bool
		expectedCode int
		expectedBody []byte
	}{
		{
			"a request with secret key should return handler's response",
			AuthConfig{
				SecretKey:    "supersecret",
				ErrorMessage: "Ah ah ah ah ahh, you didn't say the magic word",
			},
			"/?apiKey=supersecret",
			false,
			http.StatusOK,
			helloResp,
		},
		{
			"a request without secret key should return error response",
			AuthConfig{
				SecretKey:    "supersecret",
				ErrorMessage: "Ah ah ah ah ahh, you didn't say the magic word",
			},
			"/?apiKey=blah",
			true,
			http.StatusUnauthorized,
			errResponse("Ah ah ah ah ahh, you didn't say the magic word"),
		},
		{
			"default error message should be used when one is not supplied to AuthConfig",
			AuthConfig{
				SecretKey: "supersecret",
			},
			"/?apiKey=blah",
			true,
			http.StatusUnauthorized,
			errResponse("invalid API key supplied"),
		},
		{
			"supplied handler's response should be served when no secret key is set",
			AuthConfig{},
			"/",
			false,
			http.StatusOK,
			helloResp,
		},
		{
			"whitelisted routes should not require authentication",
			AuthConfig{
				SecretKey: "supersecret",
				WhiteList: []string{"/hello"},
			},
			"/hello",
			false,
			http.StatusOK,
			helloResp,
		},
	}

	for _, test := range tests {
		respRec := httptest.NewRecorder()
		handler := Auth(test.AuthConfig)(helloHandler)

		req, err := http.NewRequest(http.MethodGet, test.requestPath, nil)
		if err != nil {
			t.Fatal(err)
		}

		handler.ServeHTTP(respRec, req)

		respBody, err := ioutil.ReadAll(respRec.Body)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(test.expectedCode, respRec.Code, test.msg)
		assert.Equal(test.expectedBody, respBody, test.msg)
	}
}
