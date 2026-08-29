# Style (CONTEXT)

Recorded: 2026-08-29

## Conventions (as practiced)
- **No comments in code.** Rely on clear naming; keep only functional annotations the toolchain requires (`//go:embed`, `-- +goose Up`).
- **File naming:** snake_case per file (e.g. `user.go`, `monetary_amount.go`, `create_account_command.go`, `create_account_command_handler.go`), one concept per file. Idiomatic Go: folders/packages are single lowercase words, types are PascalCase.
- **Layering / imports:** domain → application → adapters. Domain has no deps except `google/uuid`. Application never imports adapters; adapters depend on ports (interfaces).
- **Errors:**
  - Domain: typed error structs with fields (`AccountAlreadyExistsError{OwnerID}`).
  - Application: sentinel errors (`errors.New`), compared with `errors.Is`, wrapped with `%w`.
  - Adapters: translate low-level errors to domain types (postgres `translateError`); HTTP maps sentinel → status in `error_handler.go`, never leaking internal messages on 500.
- **HTTP SRP:** thin handlers in `handler/` (bind → call use case → write response); request DTOs in `request/`, response DTOs in `response/` with `NewXResponse(...)` constructors; error mapping in a dedicated `handler/error_handler.go`.
- **Events:** past-tense names (`AccountCreated`, `MoneyDeposited`), plain structs with `uuid.UUID` + `time.Time` fields.
- **Enums:** string-typed consts (`type AccountStatus string`).
- **SQL:** statements as package-level consts in the repository file.
- **Tests:** stdlib `testing` only, hand-rolled fakes — see `.ai/context/testing.md`.

## Observed inconsistencies (FACT — decide/standardize later)
- `AccountStatus` values lowercase (`"active"`) vs `AccountTransactionStatus` UPPERCASE (`"PENDING"`).
- `OwnerId` (domain fields, `UserCreated.OwnerId`) vs Go initialism convention `OwnerID`; also `accountId` vs `accountID` in `Deposit`.
- `NewDeposit`/`NewTransfer` are structs with injected deps (unusual) — elsewhere constructors are functions (`NewAccount`).
- Money is `float64` (`MonetaryAmount`); ADR-0004 records a TODO to migrate to a decimal money type.
- Stale comments exist: root `main.go` (GoLand TIP comments — dead file) and `TODO(rfontt)` in `deposit.go`.
