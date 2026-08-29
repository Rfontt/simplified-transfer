# ADR-0006: Aggregate invariants live in the domain (factories + value objects)

Status: accepted (2026-08-29)

> Superseded in part by ADR-0007: per-field value objects for user fields were
> replaced by plain string fields + a `validateFields` method, and the 3-layer
> error chain was amended (field-constraint errors now flow domain → HTTP
> directly via the shared `domain.ConstraintValidationError`). The factory/
> thin-handler principles below remain.

## Context
`CreateUserCommandHandler` and `CreateAccountCommandHandler` inlined business
validation (empty full name, CPF/CNPJ, empty email/password, currency, balance)
and defined parallel application-level sentinel errors. This duplicated domain
knowledge in the application layer: `User`/`Account` could be constructed in an
invalid state, and every future use case touching them would have to remember to
re-validate.

## Decision
- **Aggregate factories enforce invariants**: `user.NewUser(...) (*User, error)`
  and `account.NewAccount(...) (*Account, error)` refuse invalid state.
- **Value objects own their rules**: `user.FullName`, `user.Document`,
  `user.Email`, `user.Password` (boundary value for the plaintext before
  hashing), each with a constructor returning a typed domain error.
- **Application handlers are thin**: parse boundary formats (UUID), hash via the
  `PasswordHasher` port, call the factory, persist, and *map* domain errors to
  application sentinels (`errors.go`) — they no longer inline rules.
- **3-layer error chain unchanged**: adapter → domain error → app sentinel →
  HTTP status. `errors.go` sentinels stay as the app↔HTTP contract.
- `user.ValidateDocument` free function replaced by `user.NewDocument`;
  documents are **stored normalized (digits only)** so the DB uniqueness rule
  cannot be bypassed by punctuation (`529.982.247-25` == `52998224725`).
- `ownerID` UUID parsing stays in the handler (boundary format, not a domain rule).
- `MonetaryAmount` keeps its helpers; currency/balance rules validated in
  `NewAccount` (no `NewMonetaryAmount` yet — no extra value).

## Consequences
- `User`/`Account` cannot exist in an invalid state; tests moved to the domain
  (`user_test.go`, `account_test.go`).
- `user.NewUser` signature changed (returns error, takes `userType string`).
- Supersedes the "pure domain function `ValidateDocument`" shape from ADR-0005;
  the *rule* stays in domain, just as a VO constructor.
- Postgres adapter `scanUser` now converts scanned strings into VOs explicitly.
