package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nickhstr/goweb/middleware"
	c "github.com/smartystreets/goconvey/convey"
)

func TestHeaders(t *testing.T) {
	helloResp := []byte("Hello world")
	helloHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(helloResp)
	})

	c.Convey("Given a headers config", t, func() {
		var config middleware.AppHeaders

		c.Convey("When the config is not empty", func() {
			appName := "test-app"
			appVersion := "1.0.0"
			gitRevision := "abc123"
			region := "local"
			config = middleware.AppHeaders{
				AppName:     appName,
				AppVersion:  appVersion,
				GitRevision: gitRevision,
				Region:      region,
			}
			handler := middleware.Headers(config)(helloHandler)

			c.Convey("The headers middleware should add headers from the config", func() {
				expected := map[string]string{
					http.CanonicalHeaderKey("webcakes-app-name"):    appName,
					http.CanonicalHeaderKey("webcakes-app-version"): appVersion + "-" + gitRevision,
					http.CanonicalHeaderKey("webcakes-region"):      region,
				}

				respRec := httptest.NewRecorder()

				req, err := http.NewRequest(http.MethodGet, "/", nil)
				if err != nil {
					t.Fatal(err)
				}

				handler.ServeHTTP(respRec, req)

				var matchingHeaders = true
				for key, val := range respRec.Header() {
					if val[0] != expected[key] {
						matchingHeaders = false
					}
				}

				c.So(matchingHeaders, c.ShouldBeTrue)
			})
		})
	})
}
