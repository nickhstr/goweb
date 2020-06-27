package router_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	"github.com/nickhstr/goweb/router"
	"github.com/stretchr/testify/assert"
)

func TestDefaultMux(t *testing.T) {
	type requestTest struct {
		*http.Request
		expectedStatus int
	}

	tests := []struct {
		name         string
		routes       []router.Route
		opts         router.DefaultOptions
		requestTests []requestTest
	}{
		{
			"router should respond as expected with supplied routes",
			[]router.Route{
				func(r *mux.Router) {
					r.HandleFunc("/testing", func(w http.ResponseWriter, r *http.Request) {
						fmt.Fprint(w, "aww yeah")
					}).Methods(http.MethodGet)
				},
				func(r *mux.Router) {
					r.HandleFunc("/idontexist", http.NotFound)
				},
			},
			router.DefaultOptions{
				GitCommit: "abc123",
				Name:      "test-app",
				Region:    "local",
				Version:   "1.0.0",
			},
			[]requestTest{
				{
					&http.Request{
						Method: http.MethodGet,
						URL: &url.URL{
							Path: "/testing",
						},
					},
					http.StatusOK,
				},
				{
					&http.Request{
						Method: http.MethodPost,
						URL: &url.URL{
							Path: "/testing",
						},
					},
					http.StatusMethodNotAllowed,
				},
				{
					&http.Request{
						Method: http.MethodGet,
						URL: &url.URL{
							Path: "/idontexist",
						},
					},
					http.StatusNotFound,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)
			r := router.DefaultMux(test.routes, test.opts)

			for _, reqTest := range test.requestTests {
				respRec := httptest.NewRecorder()
				r.ServeHTTP(respRec, reqTest.Request)
				assert.Equal(reqTest.expectedStatus, respRec.Result().StatusCode)
			}
		})
	}
}
