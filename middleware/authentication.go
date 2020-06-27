package middleware

import (
	"net/http"
	"regexp"

	"github.com/nickhstr/goweb/write"
)

// AuthOptions holds the necessary values for the authentication middleware
type AuthOptions struct {
	WhiteList    []string
	APIKeyName   string
	SecretKey    string
	ErrorMessage string
}

// Auth handles authenticating requests.
// Authentication is driven by an API key query parameter.
func Auth(opts AuthOptions) Middleware {
	var (
		defaultAPIKeyName   = "apiKey"
		defaultErrorMessage = "invalid API key supplied"
	)

	if opts.APIKeyName == "" {
		opts.APIKeyName = defaultAPIKeyName
	}

	if opts.ErrorMessage == "" {
		opts.ErrorMessage = defaultErrorMessage
	}

	wlRegexps := make([]*regexp.Regexp, 0, len(opts.WhiteList))

	for _, pattern := range opts.WhiteList {
		r := regexp.MustCompile(pattern)
		wlRegexps = append(wlRegexps, r)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var (
				invalidKey     = false
				whitelistRoute = false
			)

			unauthHandler := badAuthHandler(opts.ErrorMessage)

			if opts.SecretKey == "" {
				next.ServeHTTP(w, r)
				return
			}

			apiKey := r.URL.Query().Get(opts.APIKeyName)
			if apiKey == "" || apiKey != opts.SecretKey {
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
				next.ServeHTTP(w, r)
			}
		})
	}
}

func badAuthHandler(err string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		write.Error(w, err, http.StatusUnauthorized)
	})
}
