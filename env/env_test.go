package env_test

import (
	"os"
	"testing"

	"github.com/nickhstr/goweb/env"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	goEnv := "GO_ENV"
	originalVal, _ := os.LookupEnv(goEnv)
	os.Setenv(goEnv, "test")

	defer os.Setenv(goEnv, originalVal)

	tests := []struct {
		name       string
		envVar     string
		defaultVal string
		expected   string
	}{
		{
			"value should be empty when env var not set",
			"SOME_ENV_VAR",
			"",
			"",
		},
		{
			"value should be non-empty when env var not set, but defaultVal supplied",
			"SOME_ENV_VAR",
			"someValue",
			"someValue",
		},
		{
			"value should be non-empty when env var is set",
			"GO_ENV",
			"",
			"test",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)
			assert.Equal(test.expected, env.Get(test.envVar, test.defaultVal))
		})
	}
}

func TestIsProd(t *testing.T) {
	assert := assert.New(t)

	goEnv := "GO_ENV"
	originalVal := env.Get(goEnv)
	os.Setenv(goEnv, "production")

	defer os.Setenv(goEnv, originalVal)

	assert.True(env.IsProd(), "returns true when GO_ENV is set to 'production'")
}

func TestEnvVarsToValidate(t *testing.T) {
	originalGoEnv := env.Get("GO_ENV")
	originalLogLevel := env.Get("LOG_LEVEL")

	os.Setenv("GO_ENV", "test")
	os.Unsetenv("LOG_LEVEL")

	defer os.Setenv("GO_ENV", originalGoEnv)
	defer os.Setenv("LOG_LEVEL", originalLogLevel)

	tests := []struct {
		name        string
		vars        []string
		shouldError bool
	}{
		{
			"env vars should be valid when set",
			[]string{"GO_ENV"},
			false,
		},
		{
			"error should be returned when missing one or more env vars",
			[]string{"GO_ENV", "LOG_LEVEL"},
			true,
		},
		{
			"no validation should be done when no env vars are supplied",
			nil,
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)

			if test.shouldError {
				assert.Error(env.ValidateEnvVars(test.vars))
			} else {
				assert.Nil(env.ValidateEnvVars(test.vars))
			}
		})
	}
}
