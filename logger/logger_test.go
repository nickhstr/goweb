package logger_test

import (
	"testing"

	"github.com/nickhstr/goweb/logger"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNew(t *testing.T) {
	Convey("A new logger should be created with default settings", t, func() {
		log := logger.New("test")

		// The *zerolog.Event returned from log.Info() should be nil, as the
		// default level is Error when the level is not specified.
		So(log.Info(), ShouldBeNil)
		So(log.Error(), ShouldNotBeNil)
	})
}

func TestNewWithLevel(t *testing.T) {
	Convey("A new logger should log at the given level", t, func() {
		log := logger.NewWithLevel("test", "info")

		So(log.Info(), ShouldNotBeNil)
	})
}
