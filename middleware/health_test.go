package middleware

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	c "github.com/smartystreets/goconvey/convey"
)

func TestHealth(t *testing.T) {
	helloResp := []byte("Hello world")
	helloHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(helloResp)
	})

	c.Convey("Given a HealthConfig", t, func() {
		var hc HealthConfig
		cbResp := map[string]string{
			"health": "all good!",
		}

		c.Convey("When the config is valid", func() {
			hc = HealthConfig{
				Path:     "/test-health",
				Callback: func() map[string]string { return cbResp },
			}
			handler := Health(hc)(helloHandler)

			c.Convey("The callback response should be returned", func() {
				expected, err := json.Marshal(hc.Callback())
				if err != nil {
					t.Fatal(err)
				}
				respRec := httptest.NewRecorder()

				req, err := http.NewRequest(http.MethodGet, "/test-health", nil)
				if err != nil {
					t.Fatal(err)
				}

				handler.ServeHTTP(respRec, req)

				respBody, err := ioutil.ReadAll(respRec.Body)
				if err != nil {
					t.Fatal(err)
				}

				c.So(bytes.Equal(respBody, expected), c.ShouldBeTrue)
			})
		})

		c.Convey("When the config is missing Path and Callback", func() {
			hc = HealthConfig{}
			handler := Health(hc)(helloHandler)

			c.Convey("The default callback response should be returned at path '/health'", func() {
				expected, err := json.Marshal(map[string]string{})
				if err != nil {
					t.Fatal(err)
				}
				respRec := httptest.NewRecorder()

				req, err := http.NewRequest(http.MethodGet, "/health", nil)
				if err != nil {
					t.Fatal(err)
				}

				handler.ServeHTTP(respRec, req)

				respBody, err := ioutil.ReadAll(respRec.Body)
				if err != nil {
					t.Fatal(err)
				}

				c.So(bytes.Equal(respBody, expected), c.ShouldBeTrue)
			})
		})

		c.Convey("When the path is not a health path", func() {
			hc = HealthConfig{
				Path:     "/test-health",
				Callback: func() map[string]string { return cbResp },
			}
			handler := Health(hc)(helloHandler)

			c.Convey("The given handler should handle the response", func() {
				respRec := httptest.NewRecorder()

				req, err := http.NewRequest(http.MethodGet, "/", nil)
				if err != nil {
					t.Fatal(err)
				}

				handler.ServeHTTP(respRec, req)

				respBody, err := ioutil.ReadAll(respRec.Body)
				if err != nil {
					t.Fatal(err)
				}

				c.So(bytes.Equal(respBody, helloResp), c.ShouldBeTrue)
			})
		})
	})
}
