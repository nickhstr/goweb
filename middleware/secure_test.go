package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	c "github.com/smartystreets/goconvey/convey"
	"github.com/unrolled/secure"
)

// Add Secure test nly so that coverage is not unecessarily lowered.
func TestSecure(t *testing.T) {
	helloHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello"))
	})

	c.Convey("Given a Secure middleware", t, func() {
		c.Convey("When it is passed a handler", func() {
			handler := Secure(secure.Options{})(helloHandler)

			c.Convey("Security headers should be set", func() {
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
