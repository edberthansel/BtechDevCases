package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port             string
	DatabaseURL      string
	JWTSecret        string
	JWTExpiryMinutes int
	InitialBalance   int64
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:             getEnv("PORT", "8080"),
		JWTSecret:        os.Getenv("JWT_SECRET"),
		JWTExpiryMinutes: getEnvInt("JWT_EXPIRY_MINUTES", 15),
		InitialBalance:   int64(getEnvInt("INITIAL_BALANCE", 1000000)),
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	// Boleh langsung DATABASE_URL, atau dirakit dari bagian-bagiannya.
	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			getEnv("DB_USER", "btech"),
			getEnv("DB_PASSWORD", "btech_secret"),
			getEnv("DB_HOST", "localhost"),
			getEnv("DB_PORT", "5432"),
			getEnv("DB_NAME", "btech_wallet"),
		)
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}