# ADR-0007: Plain fields + validateFields instead of per-field value objects

Status: accepted (2026-08-29)

## Context
ADR-0006 introduced a value object per user field (`FullName`, `Document`,
`Email`, `Password`, `Type` enum). This exploded the codebase: every new field
with a rule would require a new type + constructor + `String()` + tests (the
`Type Type` field was a symptom). The handler still orchestrated construction.

## Decision
- **User fields are plain strings**: `FullName`, `Document`, `Email`,
  `PasswordHash`, `Type` all `string`.
- **`NewUser(id, hasher, fullName, document, email, plainPassword, userType)`
  is the single creation factory** and owns ALL creation rules:
  - `validateFields(plainPassword)` method runs every constraint in place;
    violations throw `ConstraintValidationError{Field}` with the offending
    snake_case field (`full_name`, `email`, `password`, `document`, `type`).
  - `validateDocument()` is a private method: normalizes to digits-only and
    validates CPF/CNPJ via `github.com/paemuri/brdoc` (v1.1.2) — a pure,
    maintained, public-domain library replacing ~80 lines of hand-rolled
    check-digit math (verified identical behavior).
  - Hashing happens AFTER all fields pass (fail-fast: invalid input never pays
    bcrypt) via the domain `PasswordHasher` port; an empty hash from the
    hasher is rejected (`ConstraintValidationError{Field: "password"}`).
- `COMMON`/`SHOPKEEPER` are plain string constants; `CanTransfer()` unchanged.
- Shared **`domain.ConstraintValidationError{Field}`** lives in the root
  `internal/domain` package (reused by all aggregates — user fields, account
  currency/balance, deposit amount); the 5 per-field errors are gone.
- Application handler is thin: generate ID → call factory → persist. Validation
  errors pass through untouched → HTTP 400; no sentinel mapping
  (`mapUserCreationValidationError`/`mapAccountCreationValidationError` removed).
  App sentinels remain only for infra-mapped errors (already-exists, not-found,
  invalid owner ID).

## Consequences
- No per-field types/files to maintain: adding a field = one plain field + one
  check in `validateFields`.
- All user creation rules live in `internal/domain/user/user.go`; exceptions
  in `user_exceptions.go`.
- Error chain simplified: field-constraint errors flow domain → HTTP directly
  (HTTP maps the shared type to 400); the HTTP adapter may import the root
  `domain` package but not bounded-context packages.
- Supersedes ADR-0006 in part: per-field value objects dropped; the
  "aggregate factories enforce invariants" and "handlers are thin" principles
  remain.
- Domain now imports `paemuri/brdoc` (pure validation logic, no I/O) — the
  "domain free of all but google/uuid" rule is relaxed for pure validation
  libraries (see DDD-03).
