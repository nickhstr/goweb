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
	os.Setenv("GO_ENV", "production")
	log := logger.New("test")

	tests := []struct {
		goEnv               string
		expectedNilLevel    *zerolog.Event
		expectedNotNilLevel *zerolog.Event
	}{
		{
			"production",
			log.Info(),
			log.Error(),
		},
	}

	for _, test := range tests {
		assert.Nil(test.expectedNilLevel)
		assert.NotNil(test.expectedNotNilLevel)
	}
}

func TestNewWithLevel(t *testing.T) {
	assert := assert.New(t)
	log := logger.NewWithLevel("test", "info")

	assert.NotNil(log.Info, "a new logger should log at given level")
}
