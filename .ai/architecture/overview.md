# Architecture Overview (FACTS)

Recorded: 2026-08-26 · Updated: 2026-08-29

## What it is
A simplified transfer system in Go, demonstrating DDD + CQRS + Event Sourcing. Designed via Event Storming (diagrams in `system_design/`).

## Stack
- Go 1.26.0 (module `event-driven-architecture`).
- Dependencies: `gin` (HTTP), `google/uuid` (IDs), `jackc/pgx/v5` (Postgres), `pressly/goose/v3` (migrations), `DATA-DOG/go-sqlmock` (tests). See `.ai/context/libs.md`.
- Databases: PostgreSQL wired (commands/writes). MongoDB (queries/reads) planned, NOT yet integrated (driver present in go.sum as indirect).

## Bounded Contexts
- **User** (`internal/domain/user`) — identity, type COMMON/SHOPKEEPER, authentication. Aggregate root: `User`.
- **Account** (`internal/domain/account`) — balance, deposits, transfers, status. Aggregate root: `Account` (via `Deposit`/`Transfer`).

## Core patterns
- **Event Sourcing** — immutable events are the source of truth; state reconstructed from history. (Defined, but persistence/replay NOT yet implemented.)
- **CQRS** — commands (writes) separated from queries (reads).
- **Aggregate roots** per context; **repositories as interfaces** (external implementation in adapters).
- **Domain services** (`DepositService`, `TransferService`) for cross-aggregate rules.
- **Value objects** (`MonetaryAmount`) with value equality.

## Structure
`internal/domain` (pure logic) → `internal/application` (CQRS use cases: `command/` and `query/`) → `internal/adapters` (http, postgres).

## Current status (2026-08-29)
- Account creation is the wired vertical slice: `POST /accounts` (Gin) →
  `CreateAccountCommandHandler` → Postgres `AccountRepository` (goose migrations on startup).
- Layer details live in `.ai/architecture/`: `domain.md`, `application.md`, `adapters.md`, `config.md`.
- HTTP contract source of truth: `docs/openapi.yaml` (see `.ai/context/http-api.md`).
- NOT yet implemented: transfer/deposit endpoints, MongoDB queries, event
  store persistence/replay, projections connected to the stream, authorization
  and notification adapters, user endpoints.

## Links (decision process)
- https://medium.com/@rfontt/event-storming-questionar-organizar-e-decidir-assertivamente-ab00891b6b50
- https://medium.com/@rfontt/event-storming-na-pr%C3%A1tica-2aaaf99c33fa
