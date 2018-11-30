package logger

import (
	"testing"

	"github.com/rs/zerolog"
	c "github.com/smartystreets/goconvey/convey"
)

func TestSetGlobalLevel(t *testing.T) {
	c.Convey("Given a log level", t, func() {
		var logLevel string

		c.Convey("When it is a valid log level", func() {
			logLevel = "debug"

			c.Convey("The global log level should be set", func() {
				setGlobalLevel(logLevel, "test")
				globalLevel := zerolog.GlobalLevel()

				c.So(globalLevel, c.ShouldEqual, zerolog.DebugLevel)
			})
		})

		c.Convey("When it's not a valid level", func() {
			logLevel = "invalid"

			c.Convey("The global log level should be set to the environment's level", func() {
				c.Convey("Production environment", func() {
					setGlobalLevel(logLevel, "production")
					globalLevel := zerolog.GlobalLevel()

					c.So(globalLevel, c.ShouldEqual, zerolog.ErrorLevel)
				})
				c.Convey("Development environment", func() {
					setGlobalLevel(logLevel, "development")
					globalLevel := zerolog.GlobalLevel()

					c.So(globalLevel, c.ShouldEqual, zerolog.InfoLevel)
				})
				c.Convey("Debug environment", func() {
					setGlobalLevel(logLevel, "debug")
					globalLevel := zerolog.GlobalLevel()

					c.So(globalLevel, c.ShouldEqual, zerolog.DebugLevel)
				})
			})
		})
	})
}

func TestNew(t *testing.T) {
	c.Convey("Given the need to create a new logger", t, func() {
		c.Convey("When the config is nil", func() {
			c.Convey("A new logger should be created with default settings", func() {
				logger := New(nil)

				// The *zerolog.Event returned from logger.Info() should be nil, as the
				// default level is Error when the level is not specified.
				c.So(logger.Info(), c.ShouldBeNil)
			})
		})

		c.Convey("When the config has a level specified", func() {
			config := &Config{Level: "info"}

			c.Convey("A new logger should log at the given level", func() {
				goodLog := New(config)
				// Set the global level to one at the same level or a level less than the
				// config's level. If this is not done, then a higher global level will
				// prevent lower levels from being used.
				zerolog.SetGlobalLevel(zerolog.InfoLevel)

				c.So(goodLog.Info(), c.ShouldNotBeNil)
			})
		})
	})
}
