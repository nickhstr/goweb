package env

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

func init() {
	// Only load variables from .env while in development mode
	if Dev() {
		if err := godotenv.Load(); err != nil {
			log.Warn().Err(err).Msg("Could not load .env file")
		}
	}
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

// Dev indicates if app is in dev env.
func Dev() bool {
	return Get("GO_ENV", "development") == "development"
}

// Prod indicates if app is in prod env.
func Prod() bool {
	return Get("GO_ENV", "development") == "production"
}

// ServerAddress returns an appropriate address for http.ListenAndServe to use.
func ServerAddress() string {
	port := Get("PORT", "3000")

	if Dev() {
		return fmt.Sprintf("localhost:%s", port)
	}

	return fmt.Sprintf("0.0.0.0:%s", port)
}

// ValidateEnvVars provides a way to check that a given slice of environment
// variables have been set.
func ValidateEnvVars(vars []string, shouldPanic bool) {
	if vars == nil {
		return
	}

	var (
		missingVars = []string{}
		message     string
	)

	for _, val := range vars {
		_, isSet := os.LookupEnv(val)
		if !isSet {
			missingVars = append(missingVars, val)
		}
	}

	if len(missingVars) > 0 {
		message = "Missing required env variables: " + strings.Join(missingVars, ", ")

		if shouldPanic {
			panic(message)
		} else {
			log.Warn().Msg(message)
		}
	}
}
