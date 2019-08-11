package logger

import (
	"testing"

	c "github.com/smartystreets/goconvey/convey"
)

func TestNew(t *testing.T) {
	c.Convey("Given the need to create a new logger", t, func() {
		c.Convey("A new logger should be created with default settings", func() {
			logger := New("test")

			// The *zerolog.Event returned from logger.Info() should be nil, as the
			// default level is Error when the level is not specified.
			c.So(logger.Info(), c.ShouldBeNil)
			c.So(logger.Error(), c.ShouldNotBeNil)
		})

		c.Convey("When a level specified", func() {
			c.Convey("A new logger should log at the given level", func() {
				goodLog := NewWithLevel("test", "info")

				c.So(goodLog.Info(), c.ShouldNotBeNil)
			})
		})
	})
}
