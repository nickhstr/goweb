package middleware

import (
	"crypto/md5"
	"fmt"
	"log"
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

func uptime() time.Duration {
	return time.Since(startTime)
}

// Middleware represents the function type for all middleware.
type Middleware func(http.Handler) http.Handler

// Compose adds middleware handlers to a given handler.
func Compose(handler http.Handler, middlewares ...Middleware) http.Handler {
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
		middleware = []Middleware{}
	)

	err := env.ValidateEnvVars(config.EnvVarsToValidate)
	if err != nil {
		log.Fatal(err.Error())
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
		Health(HealthConfig{
			Path: healthPath,
			Callback: func() map[string]string {
				return map[string]string{
					"name":    config.AppName,
					"version": config.AppVersion,
					"region":  config.Region,
					"sha1":    config.GitRevision,
					"uptime":  fmt.Sprintf("%vs", uptime().Seconds()),
				}
			},
		}),
		Headers(AppHeaders{
			AppName:     config.AppName,
			AppVersion:  config.AppVersion,
			GitRevision: config.GitRevision,
			Region:      config.Region,
		}),
		Secure(config.SecureOptions),
	)

	if config.Compress {
		middleware = append(middleware, handlers.CompressHandler)
	}

	middleware = append(
		middleware,
		Logger,
		Recover,
	)

	return Compose(
		config.Handler,
		middleware...,
	)
}
