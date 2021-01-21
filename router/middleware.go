package router

import (
	"crypto/md5"
	"encoding/hex"

	"github.com/gorilla/handlers"
	"github.com/newrelic/go-agent/v3/integrations/nrgorilla"
	"github.com/nickhstr/goweb/middleware"
	"github.com/nickhstr/goweb/newrelic"
	"github.com/rs/cors"
	"github.com/spf13/viper"
)

// DefaultMiddlewareOptions is the configuration struct for router's
// DefaultMiddleware.
type DefaultMiddlewareOptions struct {
	AuthOptions
	Compress  bool
	CORS      bool
	ETag      bool
	GitCommit string
	Name      string
	Region    string
	Version   string
}

// AuthOptions are options for the authentication middleware.
type AuthOptions struct {
	Enabled   bool
	WhiteList []string
}

// DefaultMiddleware provides a configurable default middleware stack for
// a router.
// This middleware is intended only to be used for routes matched by the router.
// For middleware that's not limited to the router's scope, see middleware.Compose
// for adding middleware to an http.Handler.
func DefaultMiddleware(opts DefaultMiddlewareOptions) []middleware.Middleware {
	mw := []middleware.Middleware{
		nrgorilla.Middleware(newrelic.App()),
		middleware.SecureDefault(),
		middleware.AppHeaders(middleware.AppHeadersOptions{
			GitCommit: opts.GitCommit,
			Name:      opts.Name,
			Region:    opts.Region,
			Version:   opts.Version,
		}),
	}

	if opts.AuthOptions.Enabled {
		viper.SetDefault("SECRET_KEY", "keyboard cat")
		secretHash := md5.Sum([]byte(viper.GetString("SECRET_KEY")))

		mw = append(mw, middleware.Auth(middleware.AuthOptions{
			SecretKey: hex.EncodeToString(secretHash[:]),
			WhiteList: opts.AuthOptions.WhiteList,
		}))
	}

	if opts.ETag {
		mw = append(mw, middleware.Etag)
	}

	if opts.Compress {
		mw = append(mw, handlers.CompressHandler)
	}

	if opts.CORS {
		mw = append(mw, cors.Default().Handler)
	}

	return mw
}
