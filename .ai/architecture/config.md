# Config Layer (FACTS)

Recorded: 2026-08-29

## Position
`internal/config/` — dependency-free config (stdlib `os` only), loaded once at startup.

## Shape
`Config{DatabaseURL string, HTTPPort string}`.

`Load()` reads env vars with defaults:
- `DATABASE_URL` → default `postgres://postgres:postgres@localhost:5432/simplified_transfer?sslmode=disable`
- `HTTP_PORT` → default `8080`

Helper `envOr(key, fallback)` returns the env value if non-empty, else the fallback.

## Conventions
- No config library (no viper/koanf) — keep it dependency-free for a study project.
- Defaults match `docker-compose.yml` (postgres 16, user/pass `postgres`, db `simplified_transfer`) and `.env.example`.
- Tests use `t.Setenv` to exercise defaults vs env-override paths (`config_test.go`).
