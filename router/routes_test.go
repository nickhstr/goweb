package router_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/nickhstr/goweb/router"
	"github.com/stretchr/testify/assert"
)

func TestDefaultRoutes(t *testing.T) {
	type requestTest struct {
		*http.Request
		expectedStatus int
	}

	tests := []struct {
		name string
		router.DefaultRoutesOptions
		requestsTests []requestTest
	}{
		{
			"Default routes should repond as expected",
			router.DefaultRoutesOptions{
				PrometheusOptions: router.PrometheusOptions{
					Path: "/muh-metrics",
				},
				HealthOptions: router.HealthOptions{
					GitCommit: "abc123",
					Name:      "test-app",
					Path:      "/test-app/health",
					Region:    "local",
					Version:   "1.0.0",
				},
			},
			[]requestTest{
				{
					&http.Request{
						Method: http.MethodGet,
						URL: &url.URL{
							Path: "/muh-metrics",
						},
					},
					http.StatusOK,
				},
				{
					&http.Request{
						Method: http.MethodGet,
						URL: &url.URL{
							Path: "/test-app/health",
						},
					},
					http.StatusOK,
				},
				{
					&http.Request{
						Method: http.MethodPost,
						URL: &url.URL{
							Path: "/test-app/health",
						},
					},
					http.StatusMethodNotAllowed,
				},
				{
					&http.Request{
						Method: http.MethodGet,
						URL: &url.URL{
							Path: "/debug/pprof/",
						},
					},
					http.StatusOK,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)
			r := router.New(router.DefaultRoutes(test.DefaultRoutesOptions))

			for _, rt := range test.requestsTests {
				respRec := httptest.NewRecorder()
				r.ServeHTTP(respRec, rt.Request)
				assert.Equal(rt.expectedStatus, respRec.Result().StatusCode)
			}
		})
	}
}
