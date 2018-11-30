package env_test

import (
	"os"
	"testing"

	"github.com/nickhstr/goweb/env"
	c "github.com/smartystreets/goconvey/convey"
)

func TestAppName(t *testing.T) {
	c.Convey("Given 'APP_NAME' as an environment variable", t, func() {
		os.Setenv("APP_NAME", "app-test")

		c.Convey("When no default value is supplied", func() {
			name := env.AppName()

			c.Convey("The app's name should equal the environment variable", func() {
				c.So(name, c.ShouldEqual, "app-test")
			})
		})

		c.Convey("When a default value is supplied", func() {
			name := env.AppName("default-name")

			c.Convey("The app's name should not equal the supplied default", func() {
				c.So(name, c.ShouldNotEqual, "default-name")
			})
		})
	})

	c.Convey("Given no 'APP_NAME' is set", t, func() {
		c.Convey("When no default name is supplied", func() {
			os.Unsetenv("APP_NAME")
			name := env.AppName()

			c.Convey("The app's name should equal the DefaultAppName", func() {
				c.So(name, c.ShouldEqual, env.DefaultAppName)
			})
		})

		c.Convey("When a default name is supplied", func() {
			os.Unsetenv("APP_NAME")
			name := env.AppName("default-name")

			c.Convey("The app's name should equal the default name", func() {
				c.So(name, c.ShouldEqual, "default-name")
			})
		})
	})
}
