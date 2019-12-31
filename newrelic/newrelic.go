package newrelic

import (
	"net/http"
	"os"
	"strconv"

	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/nickhstr/goweb/env"
	"github.com/nickhstr/goweb/logger"
	"github.com/rs/zerolog"
)

var (
	app *newrelic.Application
)

func init() {
	appName := env.Get("NEW_RELIC_APP_NAME")
	license := env.Get("NEW_RELIC_LICENSE_KEY")
	log := logger.New("newrelic")

	if appName == "" || license == "" {
		log.Error().
			Str("app", appName).
			Str("license", license).
			Msg("missing newrelic options")
	} else if appName != "" && license != "" {
		enabled, _ := strconv.ParseBool(env.Get("NEW_RELIC_ENABLED", "false"))

		configOptions := []newrelic.ConfigOption{
			newrelic.ConfigEnabled(enabled),
			newrelic.ConfigLicense(license),
		}

		logEnabled, _ := strconv.ParseBool(env.Get("NEW_RELIC_LOG_ENABLED", "false"))
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
	return false
}

// NewLogger returns a custom logger which satisfies the newrelic Logger interface.
func NewLogger() newrelic.Logger {
	logLevel := env.Get("NEW_RELIC_LOG_LEVEL", "error")
	log := logger.NewWithLevel("newrelic", logLevel)

	return &nrLogger{log}
}
