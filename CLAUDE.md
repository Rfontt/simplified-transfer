# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an **event-driven architecture** implementation of a simplified transfer system in Go. The project demonstrates domain-driven design (DDD) principles with a CQRS pattern (Command Query Responsibility Segregation).

The architecture was designed using event storming; see the event storming diagrams and articles referenced in the README for design context.

## Architecture: DDD + CQRS + Event-Driven

### High-Level Structure

```
internal/
├── domain/              # Core domain logic (entities, aggregates, domain events)
│   ├── user/           # User aggregate root (COMMON and SHOPKEEPER types)
│   └── account/        # Account aggregate (balance, transfers, deposits)
├── application/         # Use cases implementing CQRS
│   └── user/
│       ├── command/    # Write operations (commands and handlers)
│       └── query/      # Read operations (queries and projections)
└── adapters/           # External service integrations (notification, authorization)
```

### Key Design Patterns

1. **Aggregate Roots**: `User` and `Account` are separate aggregates with their own repositories
   - User aggregate: identity, type (COMMON/SHOPKEEPER)
   - Account aggregate: balance, transaction status

2. **Domain Events**: Events represent state changes in aggregates
   - `UserDomainEvent` (e.g., UserCreated)
   - `AccountDomainEvent` (e.g., DepositCreated, TransferCreated)
   - Events implement the `AggregateEvent[T]` interface

3. **Commands & Handlers**: Separate write logic from read logic
   - Commands: `UserDepositMoneyCommand`, `UserTransferMoneyCommand`
   - Handlers: Process commands and emit domain events

4. **Value Objects**: Immutable, domain-specific types
   - `MonetaryAmount` (currency + value)
   - `AccountTransactionStatus` enum (PENDING, COMPLETED, FAILED)

5. **Repository Pattern**: Abstract data access
   - `UserRepository`, `AccountRepository`, `DepositRepository`, `TransferRepository`

### Database Strategy

The system uses CQRS with separate databases:
- **Commands**: PostgreSQL (write operations, transactional)
- **Queries**: MongoDB (read operations, eventually consistent)

## Development Commands

### Build
```bash
go build ./cmd/simplified-transfer
```

### Test
```bash
go test ./...
go test ./internal/domain/user -v
go test ./internal/application/user/command -v
```

### Run (stub - needs implementation)
```bash
go run ./cmd/simplified-transfer/main.go
```

### Linting
```bash
go fmt ./...
```

## Code Organization Patterns

### Domain Files
- **Aggregates**: Define the core business entity (e.g., `User.go`, `Deposit.go`)
- **Services**: Business logic interfaces (e.g., `UserService.go`)
- **Events**: Domain events representing state changes (e.g., `UserDomainEvent.go`)
- **Exceptions**: Business rule violations (e.g., `UserExceptions.go`)
- **Repositories**: Data access interfaces (e.g., `UserRepository.go`)

### Application Files
- **Commands**: Request objects representing user intent
- **CommandHandlers**: Process commands, update aggregates, emit events
- **Queries**: Read model requests
- **Projections**: Build read-optimized views from domain events

### Important Constraints

1. **Functional Requirements**:
   - Users must have: full name, document (CPF/CNPJ), email, password
   - COMMON users can transfer; SHOPKEEPER users can only receive
   - All transfers must be authorized via the external gateway (https://util.devi.tools/api/v2/authorize)
   - Transfer operations must be transactional (money refunded on failure)

2. **Non-Functional Requirements**:
   - CPF/CNPJ and email must be unique
   - Users are notified on transfer receipt (SMS or email)
   - Notification service is optional; failures should use DLQ (dead-letter queue)
   - System must be available and partition-tolerant (AP in CAP theorem)

## Important Notes

- The project is in early stages with skeletal code (~330 LOC across 24 files)
- Many command handlers and query projections are stubs (empty function bodies with TODOs)
- Event handlers need implementation for event sourcing
- No HTTP/API layer is implemented yet; focus is on domain and application layers
- Dependencies: only `github.com/google/uuid v1.6.0` for ID generation

## Recent Work

The `feat/agent` branch is currently being developed with AI structure and domain events. Check commit history for context on recent changes.
