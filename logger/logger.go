// Package logger creates structured, leveled loggers.
// These loggers have a focus on performance and composability.
package logger

import (
	"io"
	"os"

	"github.com/nickhstr/goweb/config"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

var rootLogger Logger

// Logger provides a convenient alias for other packages.
type Logger = zerolog.Logger

func init() {
	rootLogger = zerolog.New(logWriter()).
		With().
		Timestamp().
		Logger()
}

func level(logLevel string) zerolog.Level {
	if logLevel == "" {
		logLevel = viper.GetString("LOG_LEVEL")
	}

	if logLevel != "" {
		level, err := zerolog.ParseLevel(logLevel)
		if err == nil {
			return level
		}
	}

	goEnv := viper.GetString("GO_ENV")

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

	if !config.IsProd() {
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
