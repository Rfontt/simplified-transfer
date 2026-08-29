# Development (CONTEXT)

Recorded: 2026-08-26 · Updated: 2026-08-29

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
- Create a user: `POST /users` `{ "full_name": "Rita", "document": "529.982.247-25", "email": "rita@example.com", "password": "secret", "type": "common" }`.

## Stack
- HTTP: Gin. Postgres: pgx (stdlib driver) + goose migrations. Tests: sqlmock (adapter error mapping).
- Library list and versions: `.ai/context/libs.md`. Code style: `.ai/context/style.md`. Tests style: `.ai/context/testing.md`.

## Code conventions
- HTTP adapters follow SRP: thin handlers in `handler/` (bind → call use case → write), request DTOs in `request/`, response DTOs in `response/`, and error→HTTP-status mapping in a dedicated `handler/error_handler.go`.
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
- Transfer/deposit command handlers are empty stubs (`internal/application/user/command`).
- No authorization/notification adapters; no MongoDB read side.
- `POST /accounts` and `POST /users` vertical slices are wired end-to-end and tested (`go test ./...` passes, 2026-08-29).
- Root `main.go` is stale GoLand boilerplate — the real entry point is `cmd/simplified-transfer/main.go`.

## Business rules (summary)
- Only COMMON initiates transfers; SHOPKEEPER only receives.
- Sufficient balance; external authorization `https://util.devi.tools/api/v2/authorize`.
- Authorization/notification failure → refund + FAILED status.
- CPF/CNPJ and email unique.
- Endpoint: `POST /transfer { value, payer, payee }`.

## Code review agent
- Trigger: `/review` — fully automated: reviews the worktree changes (`git diff HEAD` + untracked), writes findings to `.git/review-notes.json`, and opens a Hunk window with the comments already rendered. No manual step.
- Review window: **herdr** when pi runs inside herdr (`HERDR_ENV=1`) — splits a sibling pane and runs `hunk diff --agent-context .git/review-notes.json --agent-notes`; **osascript + Terminal** fallback outside herdr; chat-only if hunk is missing.
- The review applies the project ruleset `.ai/context/code-review.md` (DDD + style + process rules, each with an ID) on top of the generic `code-review` skill.
- To add a rule: append a numbered entry in `.ai/context/code-review.md` under the right category — nothing else to change.
- Hunk review skill: `hunk skill path` (bundled with `hunkdiff`, npm global). Herdr skill: `~/.pi/agent/skills/herdr/SKILL.md` (`herdr --skill`).
