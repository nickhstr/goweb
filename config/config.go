package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

func init() {
	// automatically load env vars
	viper.AddConfigPath(".")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	// setup default values
	viper.SetDefault("GO_ENV", "development")

	// ignore error
	// .env files are not necessarily required
	_ = viper.ReadInConfig()
}

// Validate provides a way to check that
// a given slice of config variables have been set.
func Validate(vars []string) error {
	var (
		missingVars []string
		err         error
	)

	if vars == nil {
		return err
	}

	for _, val := range vars {
		isSet := viper.IsSet(val)
		if !isSet {
			missingVars = append(missingVars, val)
		}
	}

	if len(missingVars) > 0 {
		err = fmt.Errorf("missing required config variables: %s", strings.Join(missingVars, ", "))
		return err
	}

	return err
}

// IsDev reports if the application is in development mode.
func IsDev() bool {
	if !viper.IsSet("GO_ENV") {
		return true
	}

	return viper.GetString("GO_ENV") == "development"
}

// IsProd reports if the application is in production mode.
func IsProd() bool {
	return viper.GetString("GO_ENV") == "production"
}
