package middleware_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/nickhstr/goweb/middleware"
)

func handler(t *testing.T, shouldCancel bool) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
			if !shouldCancel {
				t.Error("Timeout occurred before handler finished")
			}
		case <-time.After(10 * time.Millisecond):
			if shouldCancel {
				t.Error("Timmeout should have canceled")
			}
		}
	})
}

func TestTimeout(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
		handler http.HandlerFunc
	}{
		{
			"Timeout's context should cancel before handler is done",
			1 * time.Microsecond,
			handler(t, true),
		},
		{
			"Timeout's context should cancel before handler is done",
			20 * time.Millisecond,
			handler(t, false),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			h := middleware.Timeout(test.timeout)(test.handler)
			h.ServeHTTP(nil, &http.Request{})
		})
	}
}
