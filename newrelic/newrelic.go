package newrelic

import (
	"io"
	"net/http"
	"os"
	"strconv"

	nr "github.com/newrelic/go-agent"
	"github.com/nickhstr/goweb/env"
	"github.com/nickhstr/goweb/logger"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	app    nr.Application
	config nr.Config
)

func init() {
	enabled, _ := strconv.ParseBool(env.Get("NEW_RELIC_ENABLED", "false"))
	appName := env.Get("NEW_RELIC_APP_NAME")
	license := env.Get("NEW_RELIC_LICENSE_KEY")

	if !enabled {
		return
	} else if appName == "" || license == "" {
		log.Error().
			Str("app", appName).
			Str("license", license).
			Msg("Missing newrelic options")
	} else if appName != "" && license != "" {
		config = nr.NewConfig(appName, license)
		setupLog(&config)

		var err error
		if app, err = nr.NewApplication(config); err != nil {
			log.Error().
				Err(err).
				Msg("Failed to create newrelic application")
			os.Exit(1)
		}
	}
}

// App provides access to the newrelic application instance.
func App() nr.Application {
	return app
}

// Handler wraps an http.Handler with newrelic monitoring.
func Handler(h http.Handler, path string) http.Handler {
	nrApp := App()
	if nrApp != nil {
		_, handler := nr.WrapHandle(nrApp, path, h)

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

// newLogger returns a custom logger which satisfies the newrelic Logger interface.
func newLogger(w io.Writer) nr.Logger {
	logLevel := env.Get("NEW_RELIC_LOG_LEVEL", "error")

	zlog := logger.New(&logger.Config{Level: logLevel})

	return &nrLogger{zlog}
}

func setupLog(c *nr.Config) {
	logEnabled, _ := strconv.ParseBool(env.Get("NEW_RELIC_LOG_ENABLED", "false"))

	if logEnabled {
		logOut := env.Get("NEW_RELIC_LOG", "stdout")

		switch logOut {
		case "stderr":
			c.Logger = newLogger(os.Stderr)
		case "stdout":
			fallthrough
		default:
			c.Logger = newLogger(os.Stdout)
		}
	}
}
