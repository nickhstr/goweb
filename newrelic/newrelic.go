// Package newrelic allows for simple to use New Relic agent configuration.
// Standard New Relic environment variable names are used for much of
// the agent's configuration.
// Logging is done via the github.com/TheWeatherCompany/packages/go/logger
// package.
package newrelic

import (
	"net/http"
	"os"

	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/nickhstr/goweb/logger"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

var app *newrelic.Application

func init() {
	viper.SetDefault("NEW_RELIC_ENABLED", "false")
	viper.SetDefault("NEW_RELIC_LOG_ENABLED", "false")
	viper.SetDefault("NEW_RELIC_LOG_LEVEL", "error")

	enabled := viper.GetBool("NEW_RELIC_ENABLED")
	appName := viper.GetString("NEW_RELIC_APP_NAME")
	license := viper.GetString("NEW_RELIC_LICENSE_KEY")
	logEnabled := viper.GetBool("NEW_RELIC_LOG_ENABLED")
	log := logger.New("newrelic")

	configOptions := []newrelic.ConfigOption{
		newrelic.ConfigEnabled(enabled),
		newrelic.ConfigLicense(license),
		newrelic.ConfigAppName(appName),
		func(cfg *newrelic.Config) {
			cfg.ErrorCollector.RecordPanics = true
		},
	}

	if logEnabled {
		l := NewLogger()
		configOptions = append(configOptions, newrelic.ConfigLogger(l))
	}

	var err error
	if app, err = newrelic.NewApplication(configOptions...); err != nil {
		log.Error().
			Err(err).
			Msg("failed to create newrelic application")
		os.Exit(1)
	}
}

// App provides access to the newrelic application instance.
func App() *newrelic.Application {
	return app
}

// Handler wraps an http.Handler with newrelic monitoring.
func Handler(h http.Handler, path string) http.Handler {
	nrApp := App()
	if nrApp != nil {
		_, handler := newrelic.WrapHandle(nrApp, path, h)

		return handler
	}

	return h
}

type nrLogger struct {
	log zerolog.Logger
}

func (l *nrLogger) fire(e *zerolog.Event, msg string, context map[string]interface{}) {
	for key, val := range context {
		e = e.Interface(key, val)
	}

	e.Msg(msg)
}

func (l *nrLogger) Error(msg string, context map[string]interface{}) {
	logEvent := l.log.Error()
	l.fire(logEvent, msg, context)
}

func (l *nrLogger) Warn(msg string, context map[string]interface{}) {
	logEvent := l.log.Warn()
	l.fire(logEvent, msg, context)
}

func (l *nrLogger) Info(msg string, context map[string]interface{}) {
	logEvent := l.log.Info()
	l.fire(logEvent, msg, context)
}

func (l *nrLogger) Debug(msg string, context map[string]interface{}) {
	logEvent := l.log.Debug()
	l.fire(logEvent, msg, context)
}

func (l *nrLogger) DebugEnabled() bool {
	return l.log.GetLevel() == zerolog.DebugLevel
}

// NewLogger returns a custom logger which satisfies the newrelic Logger interface.
func NewLogger() newrelic.Logger {
	logLevel := viper.GetString("NEW_RELIC_LOG_LEVEL")
	log := logger.NewWithLevel("newrelic", logLevel)

	return &nrLogger{log}
}
