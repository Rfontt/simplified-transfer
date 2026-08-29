# Testing Style (CONTEXT)

Recorded: 2026-08-29

## Stack
- **stdlib `testing` only** — no testify/gomock/ginkgo.
- `github.com/DATA-DOG/go-sqlmock` for the Postgres adapter (test-only dep).
- `httptest` + `gin.SetMode(gin.TestMode)` for the HTTP adapter.

## Conventions (as practiced)
- **Hand-rolled fakes/stubs** implementing ports: `fakeAccountRepository` (in-memory map + injectable `addErr`), `stubCreateAccount` (fixed result/error). Defined in the test file of the consumer package.
- **One test function per case** with descriptive names: `TestCreateAccountCommandHandler_Success`, `TestCreateAccountCommandHandler_InvalidOwnerID`, `TestCreateAccount_Conflict`, etc. (no table-driven tests so far).
- **Assertions:** plain `if` + `t.Errorf`/`t.Fatalf` with `%v`/`%+v` in the message. Error checks use `errors.Is` / `errors.As`.
- **sqlmock patterns:** `mock.ExpectExec(regexp...)` with `WithArgs(...)`; error-path tests inject `&pgconn.PgError{Code: "23505"}` to exercise `translateError`; `sqlmock.NewRows(...).AddRow(...)` for reads; `mock.ExpectationsWereMet()` at the end.
- **HTTP tests:** build the real router with a stub use case, `httptest.NewRequest` + `httptest.NewRecorder`, assert status code and unmarshal the JSON body.
- **Compile-time interface assertions** in test files: `var _ CreateAccountUseCase = (*CreateAccountCommandHandler)(nil)` and `var _ account.AccountRepository = (*fakeAccountRepository)(nil)`.
- **Helpers:** `t.Helper()` on shared fixtures; `t.Cleanup(func(){ db.Close() })`; `t.Setenv` for config tests.

## Status (verified 2026-08-29)
- `go test ./...` passes. Coverage: domain/account, application/account/command, adapters/http, adapters/postgres, config.
- No tests yet for: user domain, user application (stubs), projections, transfer/deposit flows.
