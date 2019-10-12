package logger_test

import (
	"testing"

	"github.com/nickhstr/goweb/logger"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	assert := assert.New(t)

	log := logger.New("test")

	// The *zerolog.Event returned from log.Info() should be nil, as the
	// default level is Error when the level is not specified.
	assert.Nil(log.Info(), "info log event should be nil by default")
	assert.NotNil(log.Error(), "error log event should be non-nil by default")
}

func TestNewWithLevel(t *testing.T) {
	assert := assert.New(t)

	log := logger.NewWithLevel("test", "info")

	assert.NotNil(log.Info, "a new logger should log at given level")
}
