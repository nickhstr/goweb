package middleware

import (
	"fmt"
	"net/http"
)

// AppHeaders holds app specific header values
type AppHeaders struct {
	AppName     string
	AppVersion  string
	GitRevision string
	Region      string
}

// Headers adds more app specific headers to the response
func Headers(appHeaders AppHeaders) Middleware {
	var appVersion string

	if appHeaders.AppVersion != "" && appHeaders.GitRevision != "" {
		appVersion = fmt.Sprintf(
			"%s-%s",
			appHeaders.AppVersion,
			appHeaders.GitRevision,
		)
	}

	headers := map[string]string{
		"webcakes-app-name":    appHeaders.AppName,
		"webcakes-app-version": appVersion,
		"webcakes-region":      appHeaders.Region,
	}

	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for key, value := range headers {
				if key != "" && value != "" {
					w.Header().Set(key, value)
				}
			}

			handler.ServeHTTP(w, r)
		})
	}
}
