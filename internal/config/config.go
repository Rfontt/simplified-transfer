package config

import "os"

type Config struct {
	DatabaseURL string
	HTTPPort    string
}

const (
	defaultDatabaseURL = "postgres://postgres:postgres@localhost:5432/simplified_transfer?sslmode=disable"
	defaultHTTPPort    = "8080"
)

func Load() Config {
	return Config{
		DatabaseURL: envOr("DATABASE_URL", defaultDatabaseURL),
		HTTPPort:    envOr("HTTP_PORT", defaultHTTPPort),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
