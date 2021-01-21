package config_test

import (
	"os"
	"testing"

	"github.com/nickhstr/goweb/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestIsDev(t *testing.T) {
	assert := assert.New(t)

	goEnv := "GO_ENV"
	originalVal := viper.GetString(goEnv)
	os.Setenv(goEnv, "")

	defer os.Setenv(goEnv, originalVal)

	assert.True(config.IsDev(), "returns true when GO_ENV is not set")
}

func TestIsProd(t *testing.T) {
	assert := assert.New(t)

	goEnv := "GO_ENV"
	originalVal := viper.GetString(goEnv)
	os.Setenv(goEnv, "production")

	defer os.Setenv(goEnv, originalVal)

	assert.True(config.IsProd(), "returns true when GO_ENV is set to 'production'")
}

func TestValidate(t *testing.T) {
	originalGoEnv := viper.GetString("GO_ENV")
	originalLogLevel := viper.GetString("LOG_LEVEL")

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
				assert.Error(config.Validate(test.vars))
			} else {
				assert.Nil(config.Validate(test.vars))
			}
		})
	}
}
