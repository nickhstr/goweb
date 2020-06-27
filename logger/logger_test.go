package logger_test

import (
	"os"
	"testing"

	"github.com/nickhstr/goweb/logger"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	type levelWillLog struct {
		logEventEmitter func() *zerolog.Event
		shouldLog       bool
	}

	tests := []struct {
		name        string
		goEnv       string
		levelsToLog func(logger.Logger) []levelWillLog
	}{
		{
			"logger should log at production level",
			"production",
			func(log logger.Logger) []levelWillLog {
				return []levelWillLog{
					{
						log.Debug,
						false,
					},
					{
						log.Info,
						false,
					},
					{
						log.Warn,
						false,
					},
					{
						log.Error,
						true,
					},
				}
			},
		},
		{
			"logger should log at development level",
			"development",
			func(log logger.Logger) []levelWillLog {
				return []levelWillLog{
					{
						log.Debug,
						false,
					},
					{
						log.Info,
						true,
					},
					{
						log.Warn,
						true,
					},
					{
						log.Error,
						true,
					},
				}
			},
		},
		{
			"logger should log at debug level",
			"debug",
			func(log logger.Logger) []levelWillLog {
				return []levelWillLog{
					{
						log.Debug,
						true,
					},
					{
						log.Info,
						true,
					},
					{
						log.Warn,
						true,
					},
					{
						log.Error,
						true,
					},
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			ogEnv := os.Getenv("GO_ENV")
			os.Setenv("GO_ENV", test.goEnv)
			defer os.Setenv("GO_ENV", ogEnv)

			log := logger.New("test")
			levelsToLog := test.levelsToLog(log)

			for _, level := range levelsToLog {
				if level.shouldLog {
					assert.NotNil(level.logEventEmitter())
				} else {
					assert.Nil(level.logEventEmitter())
				}
			}
		})
	}
}

func TestNewWithLevel(t *testing.T) {
	assert := assert.New(t)
	log := logger.NewWithLevel("test", "info")

	assert.NotNil(log.Info, "a new logger should log at given level")
}
