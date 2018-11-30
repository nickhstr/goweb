package logger

import (
	"io"
	"os"

	"github.com/nickhstr/goweb/env"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// See https://golang.org/pkg/time/#pkg-constants for time layout rules
const devTimeFormat = "2006/01/2 15:04:05"

func init() {
	logLevel := env.Get("LOG_LEVEL")
	goEnv := env.Get("GO_ENV")
	// Set to empty string to use UNIX time; UNIX timestamps are shorter
	// and faster.
	zerolog.TimeFieldFormat = ""
	setGlobalLevel(logLevel, goEnv)
	log.Logger = log.Output(os.Stdout)

	if !env.Prod() {
		zerolog.TimeFieldFormat = devTimeFormat
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	}
}

func setGlobalLevel(logLevel, goEnv string) {
	level := getLevel(logLevel, goEnv)
	zerolog.SetGlobalLevel(level)
}

func getLevel(logLevel, goEnv string) zerolog.Level {
	if logLevel != "" {
		level, err := zerolog.ParseLevel(logLevel)
		if err == nil {
			return level
		}
	}

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

// Config provides options for logger.New().
type Config struct {
	Out io.Writer
	// Level allows new loggers to log at their own level.
	// However, there is one caveat: if the zerolog global level is set higher
	// than this level, log messages will be logged at the global level instead.
	Level string
}

// New creates a new zerolog.Logger.
func New(c *Config) zerolog.Logger {
	var (
		goEnv      = env.Get("GO_ENV")
		defaultOut = os.Stdout
		logger     zerolog.Logger
	)

	if c == nil {
		c = &Config{}
	}
	if c.Out == nil {
		c.Out = defaultOut
	}
	if !env.Prod() {
		c.Out = zerolog.ConsoleWriter{Out: os.Stdout}
	}

	zLevel := getLevel(c.Level, goEnv)

	logger = zerolog.
		New(c.Out).
		With().
		Timestamp().
		Logger().
		Level(zLevel)

	return logger
}
