# ADR-0004: Account creation endpoint (Gin + Postgres + goose)

Status: accepted (2026-08-26)

## Context
First vertical slice connecting HTTP → application → persistence for the
Account context. Needed an endpoint to create an Account for an already
existing User, while keeping the hexagonal (ports & adapters) layering.

## Decision
- HTTP API framework: **Gin** (`internal/adapters/http`), first HTTP dependency.
- Endpoint: `POST /accounts` with `{ owner_id, currency, balance }` in the body.
  It creates only an Account; the User is assumed to already exist.
- One account per user, enforced at the DB level via `UNIQUE(owner_id)` and a
  FK `owner_id REFERENCES users(id)`.
- Postgres driver: **jackc/pgx/v5** (stdlib `database/sql` driver) in
  `internal/adapters/postgres`, implementing the existing `AccountRepository`
  port.
- Migrations: **pressly/goose/v3** with embedded `.sql` files
  (`internal/adapters/postgres/migrations/`), applied on startup.
- Application command in `internal/application/account/command/`
  (`CreateAccountCommand` + `CreateAccountCommandHandler`), exposing a
  `CreateAccountUseCase` port consumed by the HTTP adapter.
- Error mapping: pg constraint codes 23505 → 409, 23503 → 404; application
  sentinel errors keep the HTTP adapter free of domain imports.
- Balance stored as `DOUBLE PRECISION` to match the domain `float64` (TODO:
  migrate to a decimal money type).

## Consequences
- Introduces 4 new dependencies (gin, pgx, goose, sqlmock for tests).
- `internal/adapters/` is no longer empty; Postgres and HTTP adapters now exist.
- A local Postgres (docker-compose) is required to run the service
  (`DATABASE_URL` env, with a localhost default).
- Event sourcing is still not wired: no event store, so no `AccountCreated`
  event is emitted yet.
