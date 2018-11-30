package env_test

import (
	"os"
	"testing"

	"github.com/nickhstr/goweb/env"
	c "github.com/smartystreets/goconvey/convey"
)

func TestGet(t *testing.T) {
	c.Convey("Given an environment variable to lookup", t, func() {
		envVar := "LOG_LEVEL"

		c.Convey("When the variable is not set", func() {
			os.Unsetenv(envVar)

			c.Convey("The value returned should be empty", func() {
				c.So(env.Get(envVar), c.ShouldEqual, "")
			})
		})

		c.Convey("When the variable is set", func() {
			envVarVal := "warn"
			os.Setenv(envVar, envVarVal)

			c.Convey("A value should be returned", func() {
				c.So(env.Get(envVar), c.ShouldEqual, envVarVal)
			})
		})
	})
}

func TestProd(t *testing.T) {
	goEnv := "GO_ENV"
	originalVal := env.Get(goEnv)
	defer os.Setenv(goEnv, originalVal)

	c.Convey("Given a GO_ENV variable", t, func() {
		c.Convey("When the variable is set to 'production'", func() {
			goEnvVal := "production"
			os.Setenv(goEnv, goEnvVal)

			c.Convey("It should return true", func() {
				isProd := env.Prod()
				c.So(isProd, c.ShouldEqual, true)
			})
		})
	})
}

func TestServerAddress(t *testing.T) {
	goEnv := "GO_ENV"
	originalVal := env.Get(goEnv)
	defer os.Setenv(goEnv, originalVal)

	c.Convey("Given a GO_ENV variable", t, func() {
		c.Convey("When the variable is set to 'development'", func() {
			os.Setenv(goEnv, "development")

			c.Convey("The host address should be 'localhost'", func() {
				addr := env.ServerAddress()

				c.So(addr, c.ShouldStartWith, "localhost")
			})
		})

		c.Convey("When the variable is not set to 'development'", func() {
			os.Setenv(goEnv, "production")

			c.Convey("The host address should be '0.0.0.0'", func() {
				addr := env.ServerAddress()

				c.So(addr, c.ShouldStartWith, "0.0.0.0")
			})
		})
	})
}

func TestEnvVarsToValidate(t *testing.T) {
	originalGoEnv := env.Get("GO_ENV")
	originalLogLevel := env.Get("LOG_LEVEL")
	defer os.Setenv("GO_ENV", originalGoEnv)
	defer os.Setenv("LOG_LEVEL", originalLogLevel)

	c.Convey("Given environment variables to validate", t, func() {
		vars := []string{"GO_ENV", "LOG_LEVEL"}

		c.Convey("When they are set, and shouldPanic is true", func() {
			os.Setenv("GO_ENV", "test")
			os.Setenv("LOG_LEVEL", "debug")

			c.Convey("The variables should be valid", func() {
				c.So(func() { env.ValidateEnvVars(vars, true) }, c.ShouldNotPanic)
			})
		})

		c.Convey("When one or more are not set, and shouldPanic is true", func() {
			os.Unsetenv("GO_ENV")

			c.Convey("ValidateEnvVars should panic", func() {
				c.So(func() { env.ValidateEnvVars(vars, true) }, c.ShouldPanic)
			})
		})
	})

	c.Convey("Given no variables to validate", t, func() {
		c.Convey("When a nil slice is passed to ValidateEnvVars", func() {
			c.Convey("It should do nothing", func() {
				c.So(func() { env.ValidateEnvVars(nil, true) }, c.ShouldNotPanic)
			})
		})
	})
}
