# Libraries (CONTEXT)

Recorded: 2026-08-29 — Go 1.26.0, module `event-driven-architecture`

## Direct dependencies (go.mod)
| Library | Version | Purpose | Where |
|---|---|---|---|
| `github.com/gin-gonic/gin` | v1.12.0 | HTTP router/framework (ADR-0004) | `internal/adapters/http` |
| `github.com/google/uuid` | v1.6.0 | IDs (`AggregateID` = uuid) | `internal/domain` + handlers |
| `github.com/jackc/pgx/v5` | v5.10.0 | Postgres driver (stdlib `database/sql`) | `internal/adapters/postgres` |
| `github.com/pressly/goose/v3` | v3.27.3 | SQL migrations (embedded, applied on startup) | `internal/adapters/postgres` |
| `github.com/DATA-DOG/go-sqlmock` | v1.5.2 | Postgres adapter tests (test-only) | `internal/adapters/postgres/*_test.go` |

## Notable indirects
- `go.mongodb.org/mongo-driver/v2` — present in go.sum as indirect; MongoDB read side (ADR-0002) is planned, NOT yet integrated.
- `github.com/go-playground/validator/v10` — pulled in by gin's `binding` tags used in request DTOs.

## Rules
- No new dependency without explaining why (AGENTS.md). Prefer stdlib.
- Domain layer must stay free of everything except `google/uuid`.
- Adding a library to implement a port is fine in `internal/adapters/` (e.g. gin, pgx, goose).
