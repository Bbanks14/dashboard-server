package util

import (
	"os"
	"strconv"
)

// GetEnv gets an environment variable with a fallback value
func GetEnv(key string, fallback ...string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	if len(fallback) > 0 {
		return fallback[0]
	}
	return ""
}

// GetEnvInt gets an environment variable as an integer with a fallback value
func GetEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

// GetEnvBool gets an environment variable as a boolean with a fallback value
func GetEnvBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return fallback
}
