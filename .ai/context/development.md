# Development (CONTEXT)

Recorded: 2026-08-26

## Commands
- `go test ./...` — all tests
- `go test ./internal/domain/... -v` — domain tests
- `go build ./cmd/simplified-transfer/` — build the binary
- `go fmt ./...` — format
- `go vet ./...` — static analysis
- `docker compose up -d postgres` — start local Postgres (from `docker-compose.yml`)

## Running the service
- Config is loaded in `internal/config` (`Load()`): reads env with defaults, dependency-free.
- Env: `DATABASE_URL` (defaults to `postgres://postgres:postgres@localhost:5432/simplified_transfer?sslmode=disable`) and `HTTP_PORT` (default `8080`); see `.env.example`.
- `go run ./cmd/simplified-transfer/` — applies goose migrations on startup, serves on `:8080`.
- Create an account: `POST /accounts` `{ "owner_id": "<uuid>", "currency": "BRL", "balance": 0 }`.

## Stack
- HTTP: Gin. Postgres: pgx (stdlib driver) + goose migrations. Tests: sqlmock (adapter error mapping).

## Code conventions
- Pure domain, no external dependencies (except `google/uuid`).
- Events named in past tense: `UserCreated`, `MoneyDeposited`.
- Repositories = interfaces in the domain; implementation in adapters.
- Commands fail fast; handlers do not emit partial state.
- No comments in code (Rita's preference). Rely on clear naming. Keep only functional annotations (e.g. `//go:embed`, `-- +goose Up`).

## Order when adding a feature
1. Define domain events (what happened?)
2. Aggregate methods (how does state change?)
3. Commands/handlers (who initiates?)
4. Domain services (should this be allowed?)
5. Projections (only after event sourcing is integrated)

## State / known gaps
- Event sourcing defined, but no persistence/replay.
- Command handlers mostly stubs.
- No HTTP API, no databases, no adapters (authorization/notification).
- Tests: convention is unit in the domain + in-memory/mock repos, but there are currently NO `*_test.go` files (`go test ./...` reports "no test files" in every package).

## Business rules (summary)
- Only COMMON initiates transfers; SHOPKEEPER only receives.
- Sufficient balance; external authorization `https://util.devi.tools/api/v2/authorize`.
- Authorization/notification failure → refund + FAILED status.
- CPF/CNPJ and email unique.
- Endpoint: `POST /transfer { value, payer, payee }`.
