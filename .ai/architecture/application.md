# Application Layer (FACTS)

Recorded: 2026-08-29

## Position
`internal/application/<context>/` — CQRS use cases: `command/` (writes) and `query/` (reads). Orchestrates domain; no framework/persistence imports.

## Account commands (`internal/application/account/command`)
- `CreateAccountCommand` — immutable data `{OwnerID string, Currency string, Balance float64}`.
- `CreateAccountCommandHandler` — constructor takes the `account.AccountRepository` port; implements the `CreateAccountUseCase` interface (`Handle(ctx, cmd) (*CreateAccountResult, error)`), which is the port consumed by the HTTP adapter (DIP).
- `CreateAccountResult` `{ID, OwnerID, Currency, Balance string/float64}` — application-level DTO.
- Boundary handling: ownerID parses as UUID → `ErrInvalidOwnerID`. Currency/balance
  rules live in `account.validateFields` (ADR-0007); validation errors pass through
  untouched as `domain.ConstraintValidationError` → HTTP 400.
- Sentinel errors in `errors.go`: `ErrAccountAlreadyExists`, `ErrOwnerNotFound`,
  `ErrInvalidOwnerID` (plain `errors.New`, compared with `errors.Is`).

## Error-mapping chain
1. Postgres adapter: DB error (pg code) → domain error type (`AccountAlreadyExistsError`, `OwnerNotFoundError`).
2. Command handler: infra-mapped domain errors → application sentinel, wrapped with `%w`; **validation errors pass through untouched** (no mapping).
3. HTTP adapter: sentinel → HTTP status; `domain.ConstraintValidationError` → 400 directly.

Sentinels remain only for infra-mapped errors (already-exists → 409, not-found → 404, invalid owner ID → 400); the HTTP adapter may import the root `domain` package for the shared error type.

## User commands (`internal/application/user/command`)
- `CreateUserCommand` — immutable data `{FullName, Document, Email, Password, Type string}`.
- `CreateUserCommandHandler` — constructor takes the `user.UserRepository` and `user.PasswordHasher` ports; implements the `CreateUserUseCase` interface (`Handle(ctx, cmd) (*CreateUserResult, error)`), the port consumed by the HTTP adapter.
- `CreateUserResult` `{ID, FullName, Document, Email, Type}` — application-level DTO; password hash never leaves the handler.
- Boundary handling (fail-fast): handler generates the ID and calls the domain factory
  `user.NewUser(id, hasher, ...)` — which validates ALL fields (`validateFields`),
  normalizes, then hashes. Invalid input never pays a bcrypt hash. Validation errors
  pass through untouched as `domain.ConstraintValidationError` → HTTP 400 (ADR-0007).
- Sentinel errors in `errors.go`: `ErrUserAlreadyExists` (plain `errors.New`, compared
  with `errors.Is`).
- `UserDepositMoneyCommand`, `UserTransferMoneyCommand` still exist; their handler files remain empty stubs. Not wired.

## User queries (`internal/application/user/query`)
- `UserBalanceQuery{UserID string}`, `UserBalanceQueryHandler` (depends on the `account.AccountQueries` port — read side, no adapter yet), `UserBalanceQueryProjection{UserID, Balance, Currency}`.
- Pattern: query handler returns a projection DTO; queries must have no side effects.

## Conventions
- Command/query structs are dumb data; handlers hold the logic and injected ports.
- Handlers must not emit partial state: validate everything before mutating/persisting.
- Each use case exposes a small interface (e.g. `CreateAccountUseCase`) so adapters depend on the port, not the concrete handler.
