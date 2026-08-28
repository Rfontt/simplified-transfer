package config

import "testing"

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("HTTP_PORT", "")

	cfg := Load()

	if cfg.DatabaseURL != defaultDatabaseURL {
		t.Errorf("expected default database url %q, got %q", defaultDatabaseURL, cfg.DatabaseURL)
	}
	if cfg.HTTPPort != defaultHTTPPort {
		t.Errorf("expected default http port %q, got %q", defaultHTTPPort, cfg.HTTPPort)
	}
}

func TestLoad_FromEnv(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example:5432/db")
	t.Setenv("HTTP_PORT", "9090")

	cfg := Load()

	if cfg.DatabaseURL != "postgres://example:5432/db" {
		t.Errorf("expected database url from env, got %q", cfg.DatabaseURL)
	}
	if cfg.HTTPPort != "9090" {
		t.Errorf("expected http port 9090, got %q", cfg.HTTPPort)
	}
}
