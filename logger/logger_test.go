package logger_test

import (
	"os"
	"testing"

	"github.com/rs/zerolog"

	"github.com/nickhstr/goweb/logger"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	assert := assert.New(t)

	type levelWillLog struct {
		logEventEmitter func() *zerolog.Event
		shouldLog       bool
	}

	tests := []struct {
		goEnv       string
		levelsToLog func(logger.Logger) []levelWillLog
	}{
		{
			"production",
			func(log logger.Logger) []levelWillLog {
				return []levelWillLog{
					levelWillLog{
						log.Debug,
						false,
					},
					levelWillLog{
						log.Info,
						false,
					},
					levelWillLog{
						log.Warn,
						false,
					},
					levelWillLog{
						log.Error,
						true,
					},
				}
			},
		},
		{
			"development",
			func(log logger.Logger) []levelWillLog {
				return []levelWillLog{
					levelWillLog{
						log.Debug,
						false,
					},
					levelWillLog{
						log.Info,
						true,
					},
					levelWillLog{
						log.Warn,
						true,
					},
					levelWillLog{
						log.Error,
						true,
					},
				}
			},
		},
		{
			"debug",
			func(log logger.Logger) []levelWillLog {
				return []levelWillLog{
					levelWillLog{
						log.Debug,
						true,
					},
					levelWillLog{
						log.Info,
						true,
					},
					levelWillLog{
						log.Warn,
						true,
					},
					levelWillLog{
						log.Error,
						true,
					},
				}
			},
		},
	}

	for _, test := range tests {
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
	}
}

func TestNewWithLevel(t *testing.T) {
	assert := assert.New(t)
	log := logger.NewWithLevel("test", "info")

	assert.NotNil(log.Info, "a new logger should log at given level")
}
