package middleware

import (
	"net/http"
	"net/http/pprof"
	"strings"
)

func Debug(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/debug/pprof") {
			handler.ServeHTTP(w, r)
			return
		}

		switch r.URL.Path {
		// redirect to `/debug/pprof/`
		case "/debug/pprof":
			redirPath := r.URL.Path + "/"
			if r.URL.RawQuery != "" {
				redirPath = redirPath + "?" + r.URL.RawQuery
			}

			http.Redirect(w, r, redirPath, http.StatusMovedPermanently)
		case "/debug/pprof/cmdline":
			pprof.Cmdline(w, r)
		case "/debug/pprof/profile":
			pprof.Profile(w, r)
		case "/debug/pprof/symbol":
			pprof.Symbol(w, r)
		case "/debug/pprof/trace":
			pprof.Trace(w, r)
		default:
			pprof.Index(w, r)
		}
	})
}
