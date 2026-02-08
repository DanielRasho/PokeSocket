package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

type RunningModeConfig struct {
	RunningMode string
	ApiPort     string
}

type LoggingConfig struct {
	LoggingLevel string
	PrettyLogs   string
}

type CorsConfig struct {
	AllowedOrigins      []string
	AllowedContentTypes []string
	AllowedMethods      []string
	AllowedHeaders      []string
}

// LoadConfig loads configuration from environment variables.
// It exits with a fatal error if any required variable is missing.

func LoadRunningModeConfig() RunningModeConfig {
	return RunningModeConfig{
		ApiPort:     mustGetEnv("API_PORT"),
		RunningMode: mustGetEnv("RUNNING_MODE"),
	}
}

func LoadUsersDBConfig() string {
	return mustGetEnv("USERS_DB_URI")
}

func LoadLoggingConfig() LoggingConfig {
	return LoggingConfig{
		LoggingLevel: mustGetEnv("LOGGING_LEVEL"),
		PrettyLogs:   mustGetEnv("LOGGING_PRETTY"),
	}
}

func LoadCorsConfig() CorsConfig {
	return CorsConfig{
		AllowedOrigins:      mustGetEnvAsStringSlice("ALLOWED_ORIGINS"),
		AllowedContentTypes: mustGetEnvAsStringSlice("ALLOWED_CONTENT_TYPES"),
		AllowedMethods:      mustGetEnvAsStringSlice("ALLOWED_METHODS"),
		AllowedHeaders:      mustGetEnvAsStringSlice("ALLOWED_HEADERS"),
	}
}

func GetSystemID() int {
	return mustGetEnvAsInt("SYSTEM_ID")
}

func GetJWTSecret() string {
	return mustGetEnv("JWT_SECRET")
}

func GetRunningMode() string {
	return mustGetEnvAsEnum("RUNNING_MODE", []string{"dev", "testing", "prod"})
}

// mustGetEnv retrieves the value of the given environment variable
// or exits with a fatal error if the variable is not set.
func mustGetEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatal().Msgf("Environment variable %s is required but not set", key)
	}
	return value
}

// mustGetEnvAsInt retrieves the value of the given environment variable,
// converts it to an integer, or exits with a fatal error if the variable
// is not set or cannot be converted to an integer.
func mustGetEnvAsInt(key string) int {
	valueStr := mustGetEnv(key)
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Fatal().Msgf("Environment variable %s must be a valid integer: %v", key, err)
	}
	return value
}

func mustGetEnvAsStringSlice(key string) []string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatal().Msgf("Environment variable %s is required but not set", key)
	}

	return strings.Split(value, ",")
}

func mustGetEnvAsEnum(key string, enumValues []string) string {
	value := mustGetEnv(key)
	for _, enumValue := range enumValues {
		if value == enumValue {
			return value
		}
	}
	log.Fatal().Msgf("Environment variable %s must be one of %v", key, enumValues)
	return ""
}

func getEnvOrDefault(key, def string) string {
	value, exist := os.LookupEnv(key)
	if !exist {
		return def
	}
	return value
}
