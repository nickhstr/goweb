package logger

import (
	"io"
	"os"

	"github.com/nickhstr/goweb/env"
	"github.com/rs/zerolog"
)

// Logger provides a convenient alias for other packages
type Logger = zerolog.Logger

var rootLogger Logger

func init() {
	rootLogger = zerolog.New(logWriter()).
		With().
		Timestamp().
		Logger()
}

func level(logLevel string) zerolog.Level {
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
	case "test":
		return zerolog.PanicLevel
	case "production":
		fallthrough
	default:
		return zerolog.ErrorLevel
	}
}

func logWriter() io.Writer {
	// See https://golang.org/pkg/time/#pkg-constants for time layout rules
	const devTimeFormat = "2006/01/2 15:04:05"
	var out io.Writer

	if !env.IsProd() {
		out = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: devTimeFormat}
	} else {
		out = os.Stdout
	}

	return out
}

// New creates a new child logger.
func New(namespace string) Logger {
	logger := createLogger(namespace, "")

	return logger
}

// NewWithLevel creates a new child logger with the specified level and output.
func NewWithLevel(namespace, logLevel string) Logger {
	logger := createLogger(namespace, logLevel)

	return logger
}

func createLogger(namespace, logLevel string) Logger {
	logger := rootLogger.
		Level(level(logLevel)).
		With().
		Str("namespace", namespace).
		Logger()

	return logger
}
