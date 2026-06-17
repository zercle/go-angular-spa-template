// Package config loads runtime configuration from the environment.
package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration for the server.
type Config struct {
	DatabaseURL string
	Port        string
	AppEnv      string
}

// Load reads configuration from the environment, loading a local .env file
// first when present (development convenience; ignored in production).
func Load() Config {
	_ = godotenv.Load()

	return Config{
		DatabaseURL: getenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/app?sslmode=disable"),
		Port:        getenv("PORT", "3000"),
		AppEnv:      getenv("APP_ENV", "development"),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
