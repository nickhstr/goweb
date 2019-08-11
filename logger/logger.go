package logger

import (
	"io"
	"os"

	"github.com/nickhstr/goweb/env"
	"github.com/rs/zerolog"
)

var rootLogger zerolog.Logger

func init() {
	rootLogger = zerolog.New(getOutput()).
		With().
		Timestamp().
		Logger()
}

func getLevel(logLevel string) zerolog.Level {
	if logLevel == "" {
		logLevel = env.Get("LOG_LEVEL")
	}

	if logLevel != "" {
		level, err := zerolog.ParseLevel(logLevel)
		if err == nil {
			return level
		}
	}

	goEnv := env.Get("GO_ENV")

	switch goEnv {
	case "development":
		return zerolog.InfoLevel
	case "debug":
		return zerolog.DebugLevel
	case "production":
		fallthrough
	default:
		return zerolog.ErrorLevel
	}
}

func getOutput() io.Writer {
	// See https://golang.org/pkg/time/#pkg-constants for time layout rules
	const devTimeFormat = "2006/01/2 15:04:05"
	var out io.Writer

	if !env.Prod() {
		out = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: devTimeFormat}
	} else {
		out = os.Stdout
	}

	return out
}

// New creates a new child logger.
func New(namespace string) zerolog.Logger {
	logger := createLogger(namespace, "")

	return logger
}

// NewWithLevel creates a new child logger with the specified level and output.
func NewWithLevel(namespace, logLevel string) zerolog.Logger {
	logger := createLogger(namespace, logLevel)

	return logger
}

func createLogger(namespace, logLevel string) zerolog.Logger {
	logger := rootLogger.
		Level(getLevel(logLevel)).
		With().
		Str("namespace", namespace).
		Logger()

	return logger
}
