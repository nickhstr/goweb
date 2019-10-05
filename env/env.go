package env

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func init() {
	// Try to load vars from .env file.
	_ = godotenv.Load()
}

// Get provides a way to get the value of a supplied environment
// variable. If it is not found, either the optionally supplied default
// value is returned, or an empty string.
func Get(envVar string, defaultVal ...string) string {
	if val, isSet := os.LookupEnv(envVar); isSet {
		return val
	}

	if len(defaultVal) > 0 {
		return defaultVal[0]
	}

	return ""
}

// IsDev indicates if app is in dev env.
func IsDev() bool {
	return Get("GO_ENV", "development") == "development"
}

// IsProd indicates if app is in prod env.
func IsProd() bool {
	return Get("GO_ENV", "development") == "production"
}

// ValidateEnvVars provides a way to check that a given slice of environment
// variables have been set.
func ValidateEnvVars(vars []string) error {
	var (
		missingVars []string
		err         error
	)

	if vars == nil {
		return err
	}

	for _, val := range vars {
		_, isSet := os.LookupEnv(val)
		if !isSet {
			missingVars = append(missingVars, val)
		}
	}

	if len(missingVars) > 0 {
		errMsg := fmt.Sprintf("Missing required env variables: %s\n", strings.Join(missingVars, ", "))
		err = errors.New(errMsg)

		return err
	}

	return err
}
