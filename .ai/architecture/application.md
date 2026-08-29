# Application Layer (FACTS)

Recorded: 2026-08-29

## Position
`internal/application/<context>/` — CQRS use cases: `command/` (writes) and `query/` (reads). Orchestrates domain; no framework/persistence imports.

## Account commands (`internal/application/account/command`)
- `CreateAccountCommand` — immutable data `{OwnerID string, Currency string, Balance float64}`.
- `CreateAccountCommandHandler` — constructor takes the `account.AccountRepository` port; implements the `CreateAccountUseCase` interface (`Handle(ctx, cmd) (*CreateAccountResult, error)`), which is the port consumed by the HTTP adapter (DIP).
- `CreateAccountResult` `{ID, OwnerID, Currency, Balance string/float64}` — application-level DTO.
- Fail-fast validation BEFORE persisting: ownerID parses as UUID → `ErrInvalidOwnerID`; currency non-empty → `ErrInvalidCurrency`; balance >= 0 → `ErrInvalidBalance`.
- Sentinel errors in `errors.go`: `ErrAccountAlreadyExists`, `ErrOwnerNotFound`, `ErrInvalidOwnerID`, `ErrInvalidCurrency`, `ErrInvalidBalance` (plain `errors.New`, compared with `errors.Is`).

## Error-mapping chain (three layers)
1. Postgres adapter: DB error (pg code) → domain error type (`AccountAlreadyExistsError`, `OwnerNotFoundError`).
2. Command handler: domain error type → application sentinel error, wrapped with `%w`.
3. HTTP adapter: sentinel error → HTTP status (see `.ai/context/http-api.md`).

This keeps the HTTP adapter free of domain imports and lets the handler own business-meaningful errors.

## User commands (`internal/application/user/command`) — STUBS
- `UserDepositMoneyCommand`, `UserTransferMoneyCommand` exist; the handler files contain only `package command` (empty). Not wired.

## User queries (`internal/application/user/query`)
- `UserBalanceQuery{UserID string}`, `UserBalanceQueryHandler` (depends on the `account.AccountQueries` port — read side, no adapter yet), `UserBalanceQueryProjection{UserID, Balance, Currency}`.
- Pattern: query handler returns a projection DTO; queries must have no side effects.

## Conventions
- Command/query structs are dumb data; handlers hold the logic and injected ports.
- Handlers must not emit partial state: validate everything before mutating/persisting.
- Each use case exposes a small interface (e.g. `CreateAccountUseCase`) so adapters depend on the port, not the concrete handler.
