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
	if err := Load(); err != nil {
		return defaultValue
	}
	env := os.Getenv(key)

	if strings.TrimSpace(env) == "" {
		return defaultValue
	}

	return env
}

func GetInt(key string, defaultValue int) int {
	env := os.Getenv(key)
	if strings.TrimSpace(env) == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(env)
	if err != nil {
		return defaultValue
	}

	return intValue
}

func GetBool(key string, defaultValue bool) bool {
	env := os.Getenv(key)
	if strings.TrimSpace(env) == "" {
		return defaultValue
	}

	boolValue, err := strconv.ParseBool(env)
	if err != nil {
		return defaultValue
	}

	return boolValue
}