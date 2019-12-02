package middleware

import (
	"encoding/json"
	"net/http"
)

// HealthConfig holds the information required by the Health middleware
type HealthConfig struct {
	Path     string
	Callback func() map[string]string
}

// Health provides a route to get an app's health information
func Health(config HealthConfig) Middleware {
	var (
		defaultPath     = "/health"
		defaultCallback = func() map[string]string {
			return map[string]string{}
		}
	)

	if config.Path == "" {
		config.Path = defaultPath
	}
	if config.Callback == nil {
		config.Callback = defaultCallback
	}

	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.Path != r.URL.Path {
				handler.ServeHTTP(w, r)
			} else {
				response, err := json.Marshal(config.Callback())
				if err != nil {
					response = []byte("Unable to marshal health callback")
				}

				h := healthHandler(response)
				h.ServeHTTP(w, r)
			}
		})
	}
}

func healthHandler(response []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset-UTF-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(response)
	})
}
