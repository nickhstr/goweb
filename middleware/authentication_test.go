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

func TestAuth(t *testing.T) {
	c.Convey("Given a handler in need of authentication", t, func() {
		helloResp := []byte("Hello world")
		helloHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write(helloResp)
		})

		c.Convey("When the AuthConfig is valid", func() {
			secret := "supersecret"
			errMsg := "Ah ah ah ah ahh, you didn't say the magic word"
			ac := AuthConfig{SecretKey: secret, ErrorMessage: errMsg}
			handler := Auth(ac)(helloHandler)

			c.Convey("A request with secret key should return handler's response", func() {
				respRec := httptest.NewRecorder()

				req, err := http.NewRequest(http.MethodGet, "/?apiKey="+secret, nil)
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

			c.Convey("A request without secret key should return error response", func() {
				respRec := httptest.NewRecorder()

				expected, err := json.Marshal(unauthError{errMsg})
				if err != nil {
					t.Fatal(err)
				}

				req, err := http.NewRequest(http.MethodGet, "/?apiKey=blah", nil)
				if err != nil {
					t.Fatal(err)
				}

				handler.ServeHTTP(respRec, req)

				respBody, err := ioutil.ReadAll(respRec.Body)
				if err != nil {
					t.Fatal(err)
				}

				c.So(respRec.Code, c.ShouldEqual, http.StatusUnauthorized)
				c.So(bytes.Equal(respBody, expected), c.ShouldBeTrue)
			})
		})

		c.Convey("When no error message and no secret key is provided", func() {
			ac := AuthConfig{SecretKey: "supersecret"}
			handler := Auth(ac)(helloHandler)

			c.Convey("The default message should be used", func() {
				respRec := httptest.NewRecorder()

				expected, err := json.Marshal(unauthError{"Invalid api key supplied"})
				if err != nil {
					t.Fatal(err)
				}

				req, err := http.NewRequest(http.MethodGet, "/?apiKey=blah", nil)
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

		c.Convey("When no SECRET_KEY env variable is set", func() {
			ac := AuthConfig{}
			handler := Auth(ac)(helloHandler)

			c.Convey("The error response should indicate a missing variable", func() {
				respRec := httptest.NewRecorder()

				expected, err := json.Marshal(unauthError{"Missing `SECRET_KEY` environment variable"})
				if err != nil {
					t.Fatal(err)
				}

				req, err := http.NewRequest(http.MethodGet, "/?apiKey=blah", nil)
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

		c.Convey("When a route is whitelisted", func() {
			secret := "supersecret"
			ac := AuthConfig{
				WhiteList: []string{"/hello"},
				SecretKey: secret,
			}
			handler := Auth(ac)(helloHandler)

			c.Convey("The route's handler should not need authentication", func() {
				respRec := httptest.NewRecorder()

				req, err := http.NewRequest(http.MethodGet, "/hello", nil)
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
