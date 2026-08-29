# Adapters Layer (FACTS)

Recorded: 2026-08-29

## Position
`internal/adapters/` — implementations of domain/application ports. Currently: `http/` (Gin), `postgres/` (pgx + goose), `crypto/` (bcrypt).

## Crypto (`internal/adapters/crypto`)
- `bcrypt_hasher.go` — `BcryptHasher` implements the `user.PasswordHasher` port (`Hash`), wrapping `golang.org/x/crypto/bcrypt` at default cost. Swapping the algorithm = new adapter only, domain/application untouched.

## HTTP (`internal/adapters/http`)
- `router.go` — `NewRouter(createAccount command.CreateAccountUseCase, createUser usercommand.CreateUserUseCase) *gin.Engine`; wires handlers + routes (`POST /accounts`, `POST /users`). Router takes use-case ports, not concrete handlers.
- `handler/` — package `handler`, thin HTTP handlers: `account_http_handler.go` / `user_http_handler.go` (`AccountHTTPHandler` / `UserHTTPHandler`): bind JSON → call use case → write response. No business logic.
- `request/` — request DTOs with gin `binding` tags (`create_account_request.go`, `create_user_request.go`).
- `response/` — response DTOs with a `NewXResponse(appResult)` constructor mapping the application result to the wire format (`create_account_response.go`, `create_user_response.go`).
- `handler/error_handler.go` — `writeError(ctx, err)` + `mapError(err) (int, string)`; sentinel error → HTTP status; unknown errors → 500 with a generic message (never leak internals).
- Contract source of truth: `docs/openapi.yaml` (see `.ai/context/http-api.md`).

## Postgres (`internal/adapters/postgres`)
- `connection.go` — `Open(dsn)` using the pgx stdlib driver over `database/sql` (import `_ "github.com/jackc/pgx/v5/stdlib"`), with a Ping check.
- `account_repository.go` — implements the `account.AccountRepository` port. SQL statements are package-level consts; `scanAccount` scans via a small `rowScanner` interface (works for both `*sql.Row` and `*sql.Rows`); `translateError` maps `pgconn.PgError` codes → domain error types: `23505` unique_violation → `AccountAlreadyExistsError`, `23503` foreign_key_violation → `OwnerNotFoundError`, anything else passes through.
- `user_repository.go` — implements the `user.UserRepository` port (`Add`/`ByID`/`AllUsers`); `scanUser` via `rowScanner`; `translateUserError` maps `23505` → `AlreadyExistsError`, distinguishing `users_document_key` vs `users_email_key` through `pgErr.ConstraintName` (fallback sets both fields).
- `migrations.go` — goose provider with `//go:embed migrations/*.sql`; `Migrate(ctx, db)` runs `Up` on startup (called from `cmd/simplified-transfer/main.go`).
- Migrations: `00001_create_users.sql` (unique document + email), `00002_create_accounts.sql` (`owner_id UNIQUE REFERENCES users(id)`, `balance DOUBLE PRECISION`).

## Composition
`cmd/simplified-transfer/main.go` (real entry point) wires everything: `config.Load()` → `postgres.Open` → `postgres.Migrate` → `postgres.NewAccountRepository` + `postgres.NewUserRepository` → `command.NewCreateAccountCommandHandler` + `usercommand.NewCreateUserCommandHandler(userRepo, crypto.NewBcryptHasher())` → `apphttp.NewRouter(createAccount, createUser)` → `router.Run(":" + port)`.

Note: root `main.go` is stale GoLand boilerplate (demo loop); the real entry is `cmd/simplified-transfer/main.go`.

## Tests
- Postgres adapter: `sqlmock` (regex SQL matching, `WithArgs`, `PgError` codes + constraint names).
- HTTP adapter: `httptest` + `gin.TestMode` + stub use cases.
- Crypto adapter: `bcrypt.CompareHashAndPassword` on the produced hash.
See `.ai/context/testing.md`.
