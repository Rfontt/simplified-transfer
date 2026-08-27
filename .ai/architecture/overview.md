# Architecture Overview (FACTS)

Recorded: 2026-08-26

## What it is
A simplified transfer system in Go, demonstrating DDD + CQRS + Event Sourcing. Designed via Event Storming (diagrams in `system_design/`).

## Stack
- Go 1.26.0 (module `event-driven-architecture`).
- Only external dependency: `github.com/google/uuid v1.6.0`.
- Planned databases (NOT yet integrated): PostgreSQL (commands/writes), MongoDB (queries/reads).

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

## Current status (2026-08-26)
- Account creation is the first wired vertical slice: `POST /accounts` (Gin) →
  `CreateAccountCommandHandler` → Postgres `AccountRepository`.
- `internal/adapters/` now contains: `http/` (Gin router + controller) and
  `postgres/` (repository, connection, goose migrations).
- NOT yet implemented: transfer/deposit endpoints, MongoDB queries, event
  store persistence/replay, projections connected to the stream, authorization
  and notification adapters.

## Links (decision process)
- https://medium.com/@rfontt/event-storming-questionar-organizar-e-decidir-assertivamente-ab00891b6b50
- https://medium.com/@rfontt/event-storming-na-pr%C3%A1tica-2aaaf99c33fa
