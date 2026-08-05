package helpers

import (
	"github.com/abhissng/neuron/utils/constant"
	"github.com/spf13/viper"
	"os"
	"strings"
)

// GetServiceName returns the service name from the app config or config files
func GetServiceName() string {
	return viper.GetString(constant.Service)
}

// GetEnvironment retrieves the current environment setting from various sources.
// It checks environment variables and viper configuration in order of priority.
func GetEnvironment() string {
	if os.Getenv(constant.Environment) != "" {
		return os.Getenv(constant.Environment)
	}

	if os.Getenv(constant.RunMode) != "" {
		return os.Getenv(constant.RunMode)
	}

	if viper.GetString(constant.Environment) != "" {
		return viper.GetString(constant.Environment)
	}

	return os.Getenv(constant.Environment)
}

// GetEnvironmentSlug normalizes environment names to standard slugs.
// It maps various environment name variations to consistent short forms.
func GetEnvironmentSlug(environment string) string {
	switch strings.ToLower(environment) {
	case "dev", "development":
		return "dev"
	case "test", "testing":
		return "test"
	case "staging":
		return "staging"
	case "prod", "production":
		return "prod"
	case "uat":
		return "uat"
	default:
		return "dev"
	}
}

// MustGetEnv retrieves a required environment variable or exits the program.
// If the variable is not set or empty, it logs a fatal error and exits with code 1.
func MustGetEnv(key string) string {
	value := os.Getenv(key)

	if value == "" || strings.TrimSpace(value) == "" {
		// In a real application, you would use a proper logging system (like Zap)
		// and maybe the logging's Fatal method here.
		Printf(constant.FATAL, "FATAL ERROR: Required environment variable '%s' is not set or is empty.\n", key)
	}

	return value
}
