# Domain Layer (FACTS)

Recorded: 2026-08-29 · Updated: 2026-08-29 (ADR-0006)

## Position
`internal/domain/` — pure business logic, no external deps except `github.com/google/uuid` (via `AggregateID`).

## Shared root package (`internal/domain`)
- `AggregateID` = `uuid.UUID`; `ToAggregateID(s)` parses a string.
- `MonetaryAmount` — value object `{Currency string, Value float64}`; helpers `IsPositive()`, `IsGreaterThanOrEqual()`.
- `AggregateEvent[T any]` — generic interface `Handle(event T) error`. Defined but NOT implemented by any event (events are plain structs).

## Invariants pattern (ADR-0006)
Aggregate factories enforce invariants and return `(*Aggregate, error)`; value
objects validate in their constructors. Application handlers map domain errors,
they do not inline rules.

## Bounded context: User (`internal/domain/user`)
- Aggregate: `User` `{ID, FullName, Document, Email, PasswordHash, Type}`; `CanTransfer()` = type is COMMON.
- Value objects (each with `New...` constructor + typed error + `String()`):
  `FullName` (trimmed, non-empty), `Document` (valid CPF/CNPJ, stored digits-only),
  `Email` (trimmed, non-empty), `Password` (plaintext, non-empty — whitespace-only rejected, transient boundary VO).
- `NewUser(id, fullName, document, email, passwordHash, userType)` takes **validated VOs** —
  assembles + guards hash presence only. VOs/type are validated by their own constructors
  (`NewFullName`, `NewDocument`, `NewEmail`, `ParseType`) called by the handler BEFORE hashing
  (fail-fast). `PasswordHash` field holds the **hash** (string).
- `Type` enum: `COMMON`, `SHOPKEEPER` (lowercase string values); `ParseType` parses.
- Port: `UserRepository` — `ByID(ctx, id)`, `Add(ctx, *User)`, `AllUsers(ctx)`.
- Events: `UserCreated` (plain struct). Errors: `AlreadyExistsError{Email}`,
  `InvalidFullNameError`, `InvalidDocumentError{Document}`, `InvalidEmailError`,
  `InvalidPasswordError`, `InvalidTypeError{Type}`.

## Bounded context: Account (`internal/domain/account`)
- Aggregate: `Account` `{ID, OwnerId, Balance, Status, CreatedAt}`;
  `NewAccount(id, ownerID, balance) (*Account, error)` validates currency non-empty and balance >= 0;
  `CanTransact()` (status ACTIVE), `HasBalance(amount)`.
- `AccountStatus` enum: `active`/`blocked`/`closed` (lowercase).
- `AccountTransactionStatus` enum: `PENDING`/`COMPLETED`/`FAILED` (UPPERCASE).
- Entities: `Deposit` `{ID, AccountId, Amount, Status}`, `Transfer` `{ID, From, To, Amount, Status}`.
- Ports: `AccountRepository` (ByID/Add/AllAccounts), `DepositRepository`, `TransferRepository` (note: Deposit/Transfer repos use `id string`, not a typed ID), `AccountQueries` — read-side port `OwnerBalance(ctx, ownerID uuid.UUID)`.
- Domain services: `DepositService`, `TransferService` interfaces (adapter impls not yet written).
- Unusual pattern: `NewDeposit` / `NewTransfer` are structs that hold injected deps (service + repositories) and expose `Create()` / `CreateTransfer(ctx, ...)`. Callers must wire the deps; no plain constructor exists for these.
- Events: `AccountCreated`, `AccountTransferCreated/Succeeded/Failed`, `AccountDepositCreated/Succeeded/Failed` — plain structs with `uuid.UUID` + `time.Time` fields.
- Errors: `AccountAlreadyExistsError{OwnerID}`, `OwnerNotFoundError{OwnerID}`, `InvalidCurrencyError`, `InvalidBalanceError`.

## Business rules encoded in domain
- Only COMMON users may initiate transfers (`User.CanTransfer`); SHOPKEEPER receives only.
- Sender account must be ACTIVE and have sufficient balance (`NewTransfer.CreateTransfer`).
- Deposit amount must be positive (`NewDeposit.Create`).
- New users/accounts must be valid at construction (ADR-0006).
- Not yet encoded: external authorization, refund-on-failure, unique CPF/CNPJ/email enforcement (DB-level unique exists for accounts).
