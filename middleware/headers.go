package middleware

import (
	"fmt"
	"net/http"
	"strings"
)

// AppHeaders holds app specific header values.
type AppHeadersOptions struct {
	GitCommit string
	Name      string
	Region    string
	Version   string
}

// AppHeaders adds more app-specific headers to the response.
func AppHeaders(opts AppHeadersOptions) Middleware {
	// Default to GitCommit, as that is more often than not set,
	// unlike Version
	var appVersion = opts.GitCommit

	// Don't interpolate appVersion if opts.Version is the full commit
	// sha of opts.GitCommit
	if strings.Contains(opts.Version, opts.GitCommit) {
		appVersion = opts.GitCommit
	} else if opts.Version != "" && opts.GitCommit != "" {
		appVersion = fmt.Sprintf(
			"%s-%s",
			opts.Version,
			opts.GitCommit,
		)
	}

	headers := map[string]string{
		"app-name":    opts.Name,
		"app-version": appVersion,
		"region":      opts.Region,
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for key, value := range headers {
				if key != "" && value != "" {
					w.Header().Set(key, value)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SetHeader adds header key/values to the response.
func SetHeader(key, value string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(key, value)
			next.ServeHTTP(w, r)
		})
	}
}
