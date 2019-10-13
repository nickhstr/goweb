package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unrolled/secure"
)

// Add Secure test nly so that coverage is not unecessarily lowered.
func TestSecure(t *testing.T) {
	assert := assert.New(t)

	// Set GO_ENV to test `isProd` check
	goEnv := "GO_ENV"
	originalVal, _ := os.LookupEnv(goEnv)
	os.Setenv(goEnv, "test")
	defer os.Setenv(goEnv, originalVal)

	helloHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Hello"))
	})
	handler := Secure(secure.Options{})(helloHandler)
	respRec := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	assert.NotPanics(func() { handler.ServeHTTP(respRec, req) })
}
