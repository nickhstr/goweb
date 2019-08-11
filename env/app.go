package env

import (
	"os"
)

// DefaultAppName is the app's default name, set to the "GO_ENV" environment variable
const DefaultAppName = "web-service"

// AppName returns the application's name.
// The application name can be set as an environment variable,
// or it can passed as an argument.
func AppName(defaultNames ...string) string {
	var name string
	if appName, isSet := os.LookupEnv("APP_NAME"); isSet {
		return appName
	}

	// This is done to make the default name an optional argument
	if len(defaultNames) > 0 {
		name = defaultNames[0]
	} else {
		name = DefaultAppName
	}

	_ = os.Setenv("APP_NAME", name)

	return name
}
