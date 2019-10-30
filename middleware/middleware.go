package middleware

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/handlers"
	"github.com/nickhstr/goweb/env"
	"github.com/unrolled/secure"
)

var startTime time.Time

func init() {
	startTime = time.Now()
}

// Middleware represents the function type for all middleware.
type Middleware func(http.Handler) http.Handler

// Compose adds middleware handlers to a given handler.
// Middleware handlers are ordered from first to last.
func Compose(handler http.Handler, middlewares ...Middleware) http.Handler {
	// Reverse order of middlewares so that consumers of Compose can order
	// their middlewares from first to last
	for i := len(middlewares)/2 - 1; i >= 0; i-- {
		opp := len(middlewares) - 1 - i
		middlewares[i], middlewares[opp] = middlewares[opp], middlewares[i]
	}

	for _, middleware := range middlewares {
		handler = middleware(handler)
	}

	return handler
}

// Config holds the values needed for Create()
type Config struct {
	Auth              bool
	AppName           string
	AppVersion        string
	Compress          bool
	EnvVarsToValidate []string
	GitRevision       string
	Handler           http.Handler
	Region            string
	SecureOptions     secure.Options
	WhiteList         []string
}

// Create returns the default middleware composition, useful for all general
// services.
// Middleware should be added in a specific order. If any new middleware depends on
// other middleware, the new middleware should follow afterward.
func Create(config Config) http.Handler {
	var (
		healthPath = fmt.Sprintf("/%s/health", config.AppName)
		middleware []Middleware
	)

	// Add initial middleware
	middleware = append(
		middleware,
		Logger,
		Recover,
	)

	err := env.ValidateEnvVars(config.EnvVarsToValidate)
	if err != nil {
		// Log invalid env vars and exit
		log.Fatal().Err(err).Msg("Invalid environment variables")
	}

	if config.Auth {
		secretKey := env.Get("SECRET_KEY")
		if secretKey == "" {
			secretKey = "keyboard cat"
		}

		wl := []string{"/", healthPath}
		if config.WhiteList != nil {
			wl = config.WhiteList
		}

		middleware = append(middleware, Auth(AuthConfig{
			WhiteList: wl,
			SecretKey: fmt.Sprintf("%x", md5.Sum([]byte(secretKey))),
		}))
	}

	middleware = append(
		middleware,
		Secure(config.SecureOptions),
		Health(HealthConfig{
			Path: healthPath,
			Callback: func() map[string]string {
				return map[string]string{
					"name":    config.AppName,
					"version": config.AppVersion,
					"region":  config.Region,
					"sha1":    config.GitRevision,
					"uptime":  fmt.Sprintf("%vs", time.Since(startTime).Seconds()),
				}
			},
		}),
		Headers(AppHeaders{
			AppName:     config.AppName,
			AppVersion:  config.AppVersion,
			GitRevision: config.GitRevision,
			Region:      config.Region,
		}),
	)

	if config.Compress {
		middleware = append(middleware, handlers.CompressHandler)
	}

	return Compose(
		config.Handler,
		middleware...,
	)
}
