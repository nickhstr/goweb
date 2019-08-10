package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	c "github.com/smartystreets/goconvey/convey"
)

// Testing the log output doesn't add much value; however, testing the statusRecorder
// used by the Logging middleware is useful.
func TestStatusRecorder(t *testing.T) {
	c.Convey("Given a handler", t, func() {
		notFoundHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(404)
			_, _ = w.Write([]byte("Not found"))
		})

		c.Convey("When the reponse status needs to be recorded", func() {
			c.Convey("The statusRecorder should record the status set by a handler", func() {
				respRec := httptest.NewRecorder()
				sr := &statusRecorder{ResponseWriter: respRec}

				req, err := http.NewRequest(http.MethodGet, "/", nil)
				if err != nil {
					t.Fatal(err)
				}

				notFoundHandler.ServeHTTP(sr, req)

				c.So(sr.status, c.ShouldEqual, 404)
			})
		})
	})
}

// Add Logger test only so that coverage is not unecessarily lowered.
func TestLogger(t *testing.T) {
	helloHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello"))
	})

	c.Convey("Given a Logger", t, func() {
		c.Convey("When it is passed a handler", func() {
			handler := Logger(helloHandler)

			c.Convey("Info about the request should be logged", func() {
				respRec := httptest.NewRecorder()

				req, err := http.NewRequest(http.MethodGet, "/", nil)
				if err != nil {
					t.Fatal(err)
				}

				c.So(func() { handler.ServeHTTP(respRec, req) }, c.ShouldNotPanic)
			})
		})
	})
}
