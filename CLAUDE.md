# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an **event-driven architecture** implementation of a simplified transfer system in Go. The project demonstrates domain-driven design (DDD) principles with a CQRS pattern (Command Query Responsibility Segregation).

The architecture was designed using event storming; see the event storming diagrams and articles referenced in the README for design context.

## Architecture: DDD + CQRS + Event-Driven

### Core Concepts

**Bounded Contexts**: The system is divided into two main contexts:
- **User Context** (`internal/domain/user`): Identity, user types (COMMON/SHOPKEEPER), authentication
- **Account Context** (`internal/domain/account`): Wallet balance, transactions (deposits & transfers), transaction status

Each context has its own aggregate root and repository interfaces. Events from one context may influence commands in another.

### High-Level Structure

```
internal/
├── domain/                              # Core business logic
│   ├── user/                           # User aggregate root
│   │   ├── user.go                     # Aggregate entity
│   │   ├── user_repository.go          # Repository interface
│   │   ├── user_domain_event.go        # Domain events
│   │   └── user_exceptions.go          # Business rule violations
│   ├── account/                        # Account aggregate root
│   │   ├── account.go                  # Aggregate entity
│   │   ├── deposit.go, transfer.go     # Value entities within aggregate
│   │   ├── *_repository.go             # Repository interfaces
│   │   ├── *_domain_event.go           # Domain events
│   │   ├── *_service.go                # Domain services
│   │   └── account_queries.go          # Read-side queries (stubs)
│   ├── aggregate_event.go              # Generic event interface
│   ├── aggregate_id.go                 # ID abstraction
│   └── monetary_amount.go              # Value object for amounts
├── application/                         # Use cases (CQRS)
│   └── user/
│       ├── command/                    # Write operations
│       │   ├── user_deposit_money_command.go
│       │   └── user_transfer_money_command.go
│       └── query/                      # Read operations
│           ├── user_balance_query.go
│           └── user_balance_query_projection.go
└── adapters/                            # External integrations (stub)
```

### Key Design Patterns

1. **Event Sourcing**: State changes are represented as immutable domain events
   - Events are the source of truth; entities are reconstructed from event history
   - Events implement `AggregateEvent[T]` with timestamp and aggregate ID
   - Commands trigger events; handlers apply events to aggregates

2. **Aggregate Roots**: Each bounded context has one aggregate root
   - User aggregate: manages user identity and type
   - Account aggregate: manages balance and transactions (via Deposit, Transfer entities)
   - Repositories enforce aggregate consistency boundaries

3. **Commands & Handlers**: Separate write operations from reads
   - Commands: immutable request objects (e.g., `UserDepositMoneyCommand`)
   - Handlers: validate commands against aggregate rules, emit events
   - Currently: handlers are stubs; full event sourcing not yet integrated

4. **Repositories**: Hide persistence details
   - Defined as interfaces in domain; implementations are external adapters
   - Currently: in-memory or stubbed; PostgreSQL/MongoDB planned for later

5. **Domain Services**: Business logic that spans aggregates
   - `DepositService`, `TransferService`: validate business rules
   - Called from command handlers before state changes

6. **Value Objects**: Immutable, domain-specific types
   - `MonetaryAmount`: currency + amount
   - `AccountTransactionStatus`: enum for PENDING/COMPLETED/FAILED

## Development Commands

### Test
```bash
go test ./...                           # Run all tests
go test ./internal/domain/... -v        # Run domain tests with verbose output
go test ./internal/application/... -v   # Run application tests
go test -run TestName -v                # Run a specific test
go test ./... -cover                    # Show coverage
```

### Build
```bash
go build ./cmd/simplified-transfer/     # Build the main binary
```

### Format & Lint
```bash
go fmt ./...                            # Format all Go files
go vet ./...                            # Run static analysis (optional)
```

### Run
The application is currently in early stages. Main entry point is `cmd/simplified-transfer/main.go` (implementation pending).

## Code Organization Patterns

### Domain Layer (`internal/domain/`)
The domain layer contains pure business logic with no external dependencies.

**Aggregate Files** (e.g., `user.go`, `account.go`, `transfer.go`):
- Define the entity/aggregate, its identity, and valid state transitions
- Methods represent business operations on the aggregate
- Must not directly access repositories; changes are signaled via events

**Event Files** (e.g., `user_domain_event.go`, `account_domain_event.go`):
- Immutable event types implementing `AggregateEvent[T]`
- Each event represents a fact that occurred (past tense: "UserCreated", "MoneyDeposited")
- Include all data needed to reconstruct aggregate state
- Never throw events away; they are the audit trail

**Repository Interfaces** (e.g., `user_repository.go`):
- Define abstract contract for persistence (no implementation)
- Methods: `GetOne(id)` for reads, `Save(entity)` for writes
- Implementations are external adapters (stub for now)

**Service Files** (e.g., `deposit_service.go`):
- Encapsulate business rules that span multiple aggregates
- Validate pre-conditions before state changes
- Called from command handlers, not from entities

**Exceptions/Errors** (e.g., `user_exceptions.go`):
- Domain-specific error types representing business rule violations
- Used in validation; command handlers should check and reject invalid commands

**Value Objects** (e.g., `monetary_amount.go`):
- Immutable, domain-specific types with meaningful equality
- Have no identity; equality is based on their values

### Application Layer (`internal/application/`)
Orchestrates domain logic; implements use cases via CQRS.

**Commands** (e.g., `user_deposit_money_command.go`):
- Immutable request objects representing user intent
- Contain only the data needed to perform the operation
- No business logic; just data transfer objects

**Command Handlers** (e.g., `user_deposit_money_command_handler.go`):
- Load aggregate from repository
- Validate command against domain service rules
- Call aggregate methods to trigger state changes
- Emit resulting domain events
- Save aggregate back to repository
- Note: Currently stubbed; will integrate with event store

**Queries** (e.g., `user_balance_query.go`):
- Read-only requests for data
- No side effects; executed against read models (eventually consistent)

**Projections** (e.g., `user_balance_query_projection.go`):
- Build and maintain read-optimized views from domain events
- Subscribed to event streams; update on each event
- Currently stubbed; will integrate with MongoDB later

## Business Requirements

### Functional

**User Entity**:
- Required fields: full name, document (CPF/CNPJ), email, password
- User type: COMMON (can send & receive transfers) or SHOPKEEPER (receive only)
- Each user has an associated Account with a balance

**Transfer Rules**:
- Only COMMON users can initiate transfers
- SHOPKEEPER users can only receive transfers
- Sender must have sufficient balance
- All transfers must be authorized by external gateway: `https://util.devi.tools/api/v2/authorize`
- If authorization fails or notification fails, transfer is refunded and marked FAILED

**Deposit Rules**:
- Any user can receive deposits (e.g., initial funding)
- Deposits add to account balance

### Non-Functional

- CPF/CNPJ and email must be unique across users
- Transfer operations must be ACID (transactional); on failure, money is refunded
- User notification (SMS or email) on successful transfer receipt
- Notification service may be unavailable → use DLQ (dead-letter queue) for retry
- System prioritizes **Availability** and **Partition Tolerance** (AP in CAP theorem); eventual consistency is acceptable
- Database strategy: PostgreSQL for commands (transactional), MongoDB for queries (read models)

## Important Notes

**Project Status**:
- Early-stage implementation with ~350 LOC across 27 files
- Domain and application layers are skeletal; infrastructure (persistence, HTTP, event store) is stubbed
- No database integrations yet; full CQRS (PostgreSQL + MongoDB) is planned
- No external dependencies except `google/uuid`

**Implementation Gaps** (intended for future work):
- Event sourcing: events defined but not persisted/replayed; aggregates don't reconstruct from event history
- Command handlers: mostly stubs; need to emit events and integrate with repository/event store
- Projections: query projections not connected to event stream subscribers
- Event handlers: no subscriptions to event streams for cross-context communication
- HTTP API: no web layer; tests are unit/integration only
- Adapters: notification, authorization services not implemented

**Design Decisions**:
- Repositories are interfaces in the domain; implementations are external (to avoid coupling to frameworks)
- Value objects (e.g., `MonetaryAmount`) should be compared by value, not identity
- Domain events are the source of truth; all state changes flow through events
- Commands should fail fast on invalid input; handler should not emit partial state

## Development Patterns

When adding new features:
1. **Define domain events** first (what happened?)
2. **Add aggregate methods** to trigger the event (how does state change?)
3. **Create commands & handlers** to orchestrate the change (who initiates it?)
4. **Add domain services** to validate complex rules (should this be allowed?)
5. Don't add projections until event sourcing is integrated (they need the event stream)

When testing:
- Unit tests exercise domain logic (entities, services, value objects) without persistence
- Integration tests may use in-memory implementations of repositories
- Currently no persistence layer to test; mock repositories instead
