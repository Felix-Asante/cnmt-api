package env

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)


func Load() error {
	return godotenv.Load()
}

func GetString(key string, defaultValue string) string {
	_ = godotenv.Load()

	value := os.Getenv(key)
	if strings.TrimSpace(value) == "" {
		return defaultValue
	}
	return value
}

func GetInt(key string, defaultValue int) int {
	_ = godotenv.Load()

	value := os.Getenv(key)
	if strings.TrimSpace(value) == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

func GetBool(key string, defaultValue bool) bool {
	_ = godotenv.Load()

	value := os.Getenv(key)
	if strings.TrimSpace(value) == "" {
		return defaultValue
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return boolValue
}
