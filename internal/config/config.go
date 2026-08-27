// Package config loads application configuration from the environment.
package config

import "os"

// Config holds the runtime settings for the service.
type Config struct {
	// DatabaseURL is the PostgreSQL connection string (DSN).
	DatabaseURL string
	// HTTPPort is the TCP port the HTTP server listens on.
	HTTPPort string
}

const (
	defaultDatabaseURL = "postgres://postgres:postgres@localhost:5432/simplified_transfer?sslmode=disable"
	defaultHTTPPort    = "8080"
)

// Load reads configuration from environment variables, applying defaults
// where a value is not set. It is intentionally dependency-free.
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
