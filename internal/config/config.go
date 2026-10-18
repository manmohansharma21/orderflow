package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                           string
	DatabaseURL                    string
	Environment                    string
	ReadinessTimeoutSeconds        int
	RedisEnabled                   bool
	RedisAddr                      string
	RedisPassword                  string
	RedisDB                        int
	ProductCacheTTLSeconds         int
	ReconciliationIntervalSeconds  int
	ReconciliationOlderThanMinutes int
}

func Load() (Config, error) {
	// Local .env is optional. Environment variables take precedence.
	_ = godotenv.Load()
	port := getEnv("PORT", "8080")
	environment := getEnv("APP_ENV", "development")
	databaseURL := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(databaseURL) == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	readiness, err := intEnv("READINESS_TIMEOUT_SECONDS", 3, 1, 120)
	if err != nil {
		return Config{}, err
	}
	redisDB, err := intEnv("REDIS_DB", 0, 0, 15)
	if err != nil {
		return Config{}, err
	}
	cacheTTL, err := intEnv("PRODUCT_CACHE_TTL_SECONDS", 60, 1, 86400)
	if err != nil {
		return Config{}, err
	}
	reconcileInterval, err := intEnv("RECONCILIATION_INTERVAL_SECONDS", 300, 0, 86400)
	if err != nil {
		return Config{}, err
	}
	reconcileThreshold, err := intEnv("RECONCILIATION_OLDER_THAN_MINUTES", 15, 1, 10080)
	if err != nil {
		return Config{}, err
	}
	redisEnabled, err := strconv.ParseBool(getEnv("REDIS_ENABLED", "true"))
	if err != nil {
		return Config{}, fmt.Errorf("REDIS_ENABLED must be true or false")
	}

	return Config{
		Port: port, DatabaseURL: databaseURL, Environment: environment,
		ReadinessTimeoutSeconds: readiness, RedisEnabled: redisEnabled,
		RedisAddr: getEnv("REDIS_ADDR", "localhost:6379"), RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB: redisDB, ProductCacheTTLSeconds: cacheTTL,
		ReconciliationIntervalSeconds:  reconcileInterval,
		ReconciliationOlderThanMinutes: reconcileThreshold,
	}, nil
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func intEnv(key string, fallback, min, max int) (int, error) {
	value := getEnv(key, strconv.Itoa(fallback))
	n, err := strconv.Atoi(value)
	if err != nil || n < min || n > max {
		return 0, fmt.Errorf("%s must be an integer from %d to %d", key, min, max)
	}
	return n, nil
}
