# ADR-0005: Create user endpoint + bcrypt password hashing via domain port

Status: accepted (2026-08-29)

## Context
Second vertical slice, now for the User context: `POST /users` to create a
User (full name, document CPF/CNPJ, email, password, type COMMON/SHOPKEEPER),
persisted to Postgres. Passwords must never be stored in plaintext, and the
hashing mechanism should be swappable without touching domain/application code.

## Decision
- Endpoint: `POST /users` with `{ full_name, document, email, password, type }`.
  Account creation stays a separate step (`POST /accounts`, ADR-0004); no
  cross-aggregate transaction.
- **Cryptography port in domain**: `user.PasswordHasher` interface
  (`Hash(plain string) (string, error)`) in `internal/domain/user`. Adapters
  depend on it (DIP); the handler receives it via constructor.
- **Adapter**: `internal/adapters/crypto` with `BcryptHasher`
  (`golang.org/x/crypto/bcrypt`, promoted from indirect to direct dependency).
  Swapping to argon2/scrypt later = new adapter only.
- **Document validation**: pure domain function `user.ValidateDocument` —
  CPF (11 digits) or CNPJ (14 digits) with mod-11 check digits; all-same-digit
  documents rejected; punctuation tolerated (`529.982.247-25` accepted).
  Validation is type-agnostic (COMMON may hold CNPJ in this model).
- **Password validation**: presence only (non-empty). No format/min-length rules.
- Uniqueness: DB UNIQUE(document/email) enforced; adapter distinguishes
  `users_document_key` vs `users_email_key` (pg `ConstraintName`) and maps to
  the domain `AlreadyExistsError{Document|Email}` — which was extended from
  `{Email}` only.
- Response: 201 + `{id, full_name, document, email, type}` — password hash
  never returned. Errors: 400 invalid body/fields/document/type, 409 duplicate
  document or email, 500 generic.
- No new migration: `users` table (00001) already fits; `created_at` not added
  (nothing needs it yet).

## Consequences
- `internal/application/user/command` is no longer stub-only: real
  `CreateUserCommandHandler` (+ `CreateUserUseCase` port, `CreateUserResult`).
- `internal/adapters/crypto` added; `golang.org/x/crypto` becomes direct.
- Router signature grew: `NewRouter(createAccount, createUser)`.
- Event sourcing still not wired; no `UserCreated` event emitted on create.
