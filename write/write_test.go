package write_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nickhstr/goweb/write"
	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	tests := []struct {
		name string
		err  string
		code int
	}{
		{
			"Error should write a JSON error reponse",
			"Something has gone horribly wrong",
			http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)
			respRec := httptest.NewRecorder()
			write.Error(respRec, test.err, test.code)

			expectedBody, _ := json.Marshal(write.ErrorResponse{
				test.code,
				http.StatusText(test.code),
				test.err,
			})
			// Add newline, as write.Error uses json.Encoder, which adds a newline
			expectedBody = append(expectedBody, '\n')

			body, _ := ioutil.ReadAll(respRec.Result().Body)
			assert.Equal(expectedBody, body)
		})
	}
}

func TestEncodeJSON(t *testing.T) {
	tests := []struct {
		name      string
		v         interface{}
		shouldErr bool
	}{
		{
			"EncodeJSON should write a JSON reponse",
			struct {
				Text string `json:"text"`
			}{
				"hi there",
			},
			false,
		},
		{
			"EncodeJSON should return an error JSON response",
			make(chan int),
			true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)
			respRec := httptest.NewRecorder()

			write.EncodeJSON(respRec, test.v)

			if test.shouldErr {
				errResp := new(write.ErrorResponse)
				body, _ := ioutil.ReadAll(respRec.Result().Body)
				_ = json.Unmarshal(body, errResp)

				assert.Equal(errResp.Status, http.StatusInternalServerError)
			}
		})
	}
}
