package middleware

import (
	"encoding/json"
	"net/http"
	"regexp"
)

// AuthConfig holds the necessary values for the authentication middleware
type AuthConfig struct {
	WhiteList    []string
	APIKeyName   string
	SecretKey    string
	ErrorMessage string
}

// Auth handles authenticating the request
func Auth(config AuthConfig) Middleware {
	var (
		defaultAPIKeyName   = "apiKey"
		defaultErrorMessage = "invalid API key supplied"
	)

	if config.APIKeyName == "" {
		config.APIKeyName = defaultAPIKeyName
	}
	if config.ErrorMessage == "" {
		config.ErrorMessage = defaultErrorMessage
	}

	wlRegexps := make([]*regexp.Regexp, 0, len(config.WhiteList))
	for _, pattern := range config.WhiteList {
		r := regexp.MustCompile(pattern)
		wlRegexps = append(wlRegexps, r)
	}

	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var (
				invalidKey     = false
				whitelistRoute = false
			)

			unauthHandler := badAuthHandler(config.ErrorMessage)

			if config.SecretKey == "" {
				handler.ServeHTTP(w, r)
				return
			}

			apiKey := r.URL.Query().Get(config.APIKeyName)
			if apiKey == "" || apiKey != config.SecretKey {
				invalidKey = true
			}

			for _, re := range wlRegexps {
				if re.MatchString(r.URL.Path) {
					whitelistRoute = true
					break
				}
			}

			if invalidKey && !whitelistRoute {
				unauthHandler.ServeHTTP(w, r)
			} else {
				handler.ServeHTTP(w, r)
			}
		})
	}
}

type unauthError struct {
	Error string `json:"error"`
}

func badAuthHandler(errMsg string) http.Handler {
	errResponse, err := json.Marshal(unauthError{errMsg})
	if err != nil {
		errResponse = []byte("Unable to marshal error message")
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write(errResponse)
	})
}
