# Domain Layer (FACTS)

Recorded: 2026-08-29

## Position
`internal/domain/` — pure business logic, no external deps except `github.com/google/uuid` (via `AggregateID`).

## Shared root package (`internal/domain`)
- `AggregateID` = `uuid.UUID`; `ToAggregateID(s)` parses a string.
- `MonetaryAmount` — value object `{Currency string, Value float64}`; helpers `IsPositive()`, `IsGreaterThanOrEqual()`.
- `AggregateEvent[T any]` — generic interface `Handle(event T) error`. Defined but NOT implemented by any event (events are plain structs).

## Bounded context: User (`internal/domain/user`)
- Aggregate: `User` `{ID, FullName, Document, Email, Password, Type}`; `CanTransfer()` = type is COMMON.
- `Type` enum: `COMMON`, `SHOPKEEPER` (lowercase string values).
- Port: `UserRepository` — `ByID(ctx, id)`, `Add(ctx, *User)`, `AllUsers(ctx)`. No adapter implementation yet.
- Events: `UserCreated` (plain struct). Errors: `AlreadyExistsError{Email}`.

## Bounded context: Account (`internal/domain/account`)
- Aggregate: `Account` `{ID, OwnerId, Balance, Status, CreatedAt}`; constructor `NewAccount(...)`; `CanTransact()` (status ACTIVE), `HasBalance(amount)`.
- `AccountStatus` enum: `active`/`blocked`/`closed` (lowercase).
- `AccountTransactionStatus` enum: `PENDING`/`COMPLETED`/`FAILED` (UPPERCASE).
- Entities: `Deposit` `{ID, AccountId, Amount, Status}`, `Transfer` `{ID, From, To, Amount, Status}`.
- Ports: `AccountRepository` (ByID/Add/AllAccounts), `DepositRepository`, `TransferRepository` (note: Deposit/Transfer repos use `id string`, not a typed ID), `AccountQueries` — read-side port `OwnerBalance(ctx, ownerID uuid.UUID)`.
- Domain services: `DepositService`, `TransferService` interfaces (adapter impls not yet written).
- Unusual pattern: `NewDeposit` / `NewTransfer` are structs that hold injected deps (service + repositories) and expose `Create()` / `CreateTransfer(ctx, ...)`. Callers must wire the deps; no plain constructor exists for these.
- Events: `AccountCreated`, `AccountTransferCreated/Succeeded/Failed`, `AccountDepositCreated/Succeeded/Failed` — plain structs with `uuid.UUID` + `time.Time` fields.
- Errors: `AccountAlreadyExistsError{OwnerID}`, `OwnerNotFoundError{OwnerID}`.

## Business rules encoded in domain
- Only COMMON users may initiate transfers (`User.CanTransfer`); SHOPKEEPER receives only.
- Sender account must be ACTIVE and have sufficient balance (`NewTransfer.CreateTransfer`).
- Deposit amount must be positive (`NewDeposit.Create`).
- Not yet encoded: external authorization, refund-on-failure, unique CPF/CNPJ/email enforcement (DB-level unique exists for accounts).
