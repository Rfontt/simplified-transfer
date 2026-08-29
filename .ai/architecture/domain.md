# Domain Layer (FACTS)

Recorded: 2026-08-29 · Updated: 2026-08-29 (ADR-0006)

## Position
`internal/domain/` — pure business logic; external deps limited to `github.com/google/uuid` (via `AggregateID`) and pure validation libs (`paemuri/brdoc`).

## Shared root package (`internal/domain`)
- `AggregateID` = `uuid.UUID`; `ToAggregateID(s)` parses a string.
- `MonetaryAmount` — value object `{Currency string, Value float64}`; helpers `IsPositive()`, `IsGreaterThanOrEqual()`.
- `ConstraintValidationError{Field string}` — shared error for all aggregate field-constraint violations (`Field` is snake_case); HTTP maps it to 400 directly.
- `AggregateEvent[T any]` — generic interface `Handle(event T) error`. Defined but NOT implemented by any event (events are plain structs).

## Invariants pattern (ADR-0006, evolved by ADR-0007)
Aggregate factories enforce invariants and return `(*Aggregate, error)`; all field
rules live in a single `validateFields` method on the aggregate, violations thrown
as the shared `domain.ConstraintValidationError{Field}`. Handlers do not map or
inline these rules — the error flows through to HTTP (400).

## Bounded context: User (`internal/domain/user`)
- Aggregate: `User` `{ID, FullName, Document, Email, PasswordHash, Type}` — all plain
  strings (ADR-0007); `CanTransfer()` = type is COMMON.
- `NewUser(id, hasher, fullName, document, email, plainPassword, userType)` is the single
  creation factory owning ALL rules: `validateFields(hasher, plainPassword)` checks each
  field (trim/non-empty for full_name/email/password, document normalized digits-only +
  CPF/CNPJ via `paemuri/brdoc`, type ∈ {common, shopkeeper}), throwing
  the shared `domain.ConstraintValidationError{Field}` (snake_case), then calls the
  private
  `hashPassword(hasher, plainPassword)` which hashes via the `PasswordHasher` port
  (fail-fast: invalid input never pays bcrypt) and rejects an empty hash.
  `PasswordHash` holds the **hash** (string); plaintext never stored.
- `validateDocument()` and `hashPassword()` — private methods on `User`; document
  returns normalized digits or error, hashPassword assigns `PasswordHash`.
- All creation rules live in `user.go`; errors in `user_exceptions.go`.
- `Type`: string constants `COMMON`/`SHOPKEEPER` (lowercase values).
- Port: `UserRepository` — `ByID(ctx, id)`, `Add(ctx, *User)`, `AllUsers(ctx)`.
- Events: `UserCreated` (plain struct). Errors: `AlreadyExistsError{Email, Document}`,
  plus the shared `domain.ConstraintValidationError` for field violations.

## Bounded context: Account (`internal/domain/account`)
- Aggregate: `Account` `{ID, OwnerId, Balance, Status, CreatedAt}`;
  `NewAccount(id, ownerID, balance)` builds the struct and calls the private
  `validateFields()` — currency trimmed/non-empty and balance >= 0, violations thrown
  as `domain.ConstraintValidationError{Field: "currency"/"balance"}`;
  `CanTransact()` (status ACTIVE), `HasBalance(amount)`.
- `AccountStatus` enum: `active`/`blocked`/`closed` (lowercase).
- `AccountTransactionStatus` enum: `PENDING`/`COMPLETED`/`FAILED` (UPPERCASE).
- Entities: `Deposit` `{ID, AccountId, Amount, Status}`, `Transfer` `{ID, From, To, Amount, Status}`.
- Ports: `AccountRepository` (ByID/Add/AllAccounts), `DepositRepository`, `TransferRepository` (note: Deposit/Transfer repos use `id string`, not a typed ID), `AccountQueries` — read-side port `OwnerBalance(ctx, ownerID uuid.UUID)`.
- Domain services: `DepositService`, `TransferService` interfaces (adapter impls not yet written).
- Unusual pattern: `NewDeposit` / `NewTransfer` are structs that hold injected deps (service + repositories) and expose `Create()` / `CreateTransfer(ctx, ...)`. Callers must wire the deps; no plain constructor exists for these.
- Events: `AccountCreated`, `AccountTransferCreated/Succeeded/Failed`, `AccountDepositCreated/Succeeded/Failed` — plain structs with `uuid.UUID` + `time.Time` fields.
- Errors: `AccountAlreadyExistsError{OwnerID}`, `OwnerNotFoundError{OwnerID}`, plus the
  shared `domain.ConstraintValidationError` for field violations.

## Business rules encoded in domain
- Only COMMON users may initiate transfers (`User.CanTransfer`); SHOPKEEPER receives only.
- Sender account must be ACTIVE and have sufficient balance (`NewTransfer.CreateTransfer`).
- Deposit amount must be positive (`NewDeposit.Create`) →
  `domain.ConstraintValidationError{Field: "amount"}`.
- New users/accounts must be valid at construction (ADR-0006/0007).
- Not yet encoded: external authorization, refund-on-failure, unique CPF/CNPJ/email enforcement (DB-level unique exists for accounts).
