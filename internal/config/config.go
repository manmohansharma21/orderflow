package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                    string
	DatabaseURL             string
	Environment             string
	ReadinessTimeoutSeconds int
}

func Load() (Config, error) {
	// Local .env is optional. Environment variables take precedence.
	_ = godotenv.Load()

	port := getEnv("PORT", "8080")
	environment := getEnv("APP_ENV", "development")
	databaseURL := os.Getenv("DATABASE_URL")

	timeoutValue := getEnv("READINESS_TIMEOUT_SECONDS", "3")
	timeoutSeconds, err := strconv.Atoi(timeoutValue)
	if err != nil || timeoutSeconds < 1 {
		return Config{}, fmt.Errorf(
			"READINESS_TIMEOUT_SECONDS must be a positive integer",
		)
	}

	if strings.TrimSpace(databaseURL) == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	return Config{
		Port:                    port,
		DatabaseURL:             databaseURL,
		Environment:             environment,
		ReadinessTimeoutSeconds: timeoutSeconds,
	}, nil
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}
