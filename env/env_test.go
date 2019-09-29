package env_test

import (
	"os"
	"testing"

	"github.com/nickhstr/goweb/env"
	. "github.com/smartystreets/goconvey/convey"
)

func TestGet(t *testing.T) {
	Convey("Given an environment variable to lookup", t, func() {
		envVar := "LOG_LEVEL"

		Convey("When the variable is not set", func() {
			os.Unsetenv(envVar)

			Convey("The value returned should be empty", func() {
				So(env.Get(envVar), ShouldEqual, "")
			})
		})

		Convey("When the variable is set", func() {
			envVarVal := "warn"
			os.Setenv(envVar, envVarVal)

			Convey("A value should be returned", func() {
				So(env.Get(envVar), ShouldEqual, envVarVal)
			})
		})
	})
}

func TestIsProd(t *testing.T) {
	goEnv := "GO_ENV"
	originalVal := env.Get(goEnv)
	defer os.Setenv(goEnv, originalVal)

	Convey("Given a GO_ENV variable", t, func() {
		Convey("When the variable is set to 'production'", func() {
			goEnvVal := "production"
			os.Setenv(goEnv, goEnvVal)

			Convey("It should return true", func() {
				isProd := env.IsProd()
				So(isProd, ShouldEqual, true)
			})
		})
	})
}

func TestEnvVarsToValidate(t *testing.T) {
	originalGoEnv := env.Get("GO_ENV")
	originalLogLevel := env.Get("LOG_LEVEL")
	defer os.Setenv("GO_ENV", originalGoEnv)
	defer os.Setenv("LOG_LEVEL", originalLogLevel)

	Convey("Given environment variables to validate", t, func() {
		vars := []string{"GO_ENV", "LOG_LEVEL"}

		Convey("When they are set", func() {
			os.Setenv("GO_ENV", "test")
			os.Setenv("LOG_LEVEL", "debug")

			Convey("The variables should be valid", func() {
				So(env.ValidateEnvVars(vars), ShouldBeNil)
			})
		})

		Convey("When one or more are not set", func() {
			os.Unsetenv("GO_ENV")

			Convey("ValidateEnvVars should return an error", func() {
				So(env.ValidateEnvVars(vars), ShouldBeError)
			})
		})
	})

	Convey("Given no variables to validate", t, func() {
		Convey("When a nil slice is passed to ValidateEnvVars", func() {
			Convey("It should do nothing", func() {
				So(env.ValidateEnvVars(nil), ShouldBeNil)
			})
		})
	})
}
