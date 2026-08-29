# Code Review Ruleset (CONTEXT)

Recorded: 2026-08-29

Review agent rules for this project. The `/review` prompt and any review flow
MUST load this file and apply every rule to the reviewed diff. Comments on
findings reference rule IDs (`[DDD-01]`) so rules can evolve without breaking
references.

## How to use
1. Load this file + the `code-review` skill before reviewing.
2. Review the diff against every rule below; skip none.
3. Report findings by severity: **blocker** (must fix before merge) / **important** (should fix) / **suggestion** (nice to have).
4. Comment format: `[<RULE-ID>] <finding> — <where> — <concrete suggestion>`.
5. Never claim a rule passes without verifying the code (run `go test`, `go vet`, `go fmt` when needed).
6. To add a rule: append a new numbered entry under the right category — no other changes needed.

## DDD rules

- **[DDD-01] Invariants live in the domain aggregate.** Flag business-rule
  validation inlined in the application layer (command/query handlers). Rules
  must be enforced by aggregate factories (`NewUser`, `NewAccount`) or value
  object constructors, which refuse invalid state.
  - Anti-example (the bug that motivated this rule): `if cmd.Email == "" {
    return nil, ErrInvalidEmail }` inside `CreateUserCommandHandler` — the
    handler re-implemented domain knowledge and `User` could be constructed
    invalid.
  - Correct: `user.NewUser(...) (*User, error)` validates everything; handler
    only maps domain errors.
- **[DDD-02] Application handlers are thin orchestration.** Flag handlers that
  do more than: parse boundary formats (UUID), call ports, call the domain
  factory/aggregate, persist, and map errors. No business logic, no validation,
  no formatting rules.
- **[DDD-03] Domain layer is pure.** Flag any import in `internal/domain`
  beyond `github.com/google/uuid` — no framework, no persistence, no HTTP, no
  crypto implementation. Infrastructure (hashers, DB, HTTP) lives in adapters,
  behind ports defined in the domain.
- **[DDD-04] Concepts with rules are value objects, not raw strings.** Flag
  domain concepts that carry behavior/validation as raw `string` fields —
  `Document` (CPF/CNPJ), `Email`, `MonetaryAmount`, statuses/enums. VOs
  validate in their constructors and are immutable.
- **[DDD-05] 3-layer error chain respected.** Adapter maps infra errors → domain
  error type → application sentinel (`errors.go`) → HTTP status. Flag the HTTP
  adapter importing domain, or application sentinels duplicating rules instead
  of mapping them.
- **[DDD-06] Events are past-tense and the source of truth.** Flag event names
  not in past tense (`UserCreate`), or state changes that bypass events in an
  event-sourced aggregate.
- **[DDD-07] No cross-aggregate transactions.** Contexts communicate via events
  or eventual consistency; flag handlers mutating two aggregates in one
  transaction.

## Style rules (project conventions)

- **[STYLE-01] No comments in code.** Flag explanatory comments; only
  functional annotations required by the toolchain are allowed (`//go:embed`,
  `-- +goose Up`). Rely on clear naming.
- **[STYLE-02] HTTP adapters follow SRP.** Flag fat controllers. Controllers
  bind → call use case → write response; request DTOs in `request/`, response
  DTOs in `response/` (constructor maps app result → DTO); error→HTTP-status
  mapping only in `handler/error_handler.go`.
- **[STYLE-03] Verified claims.** Flag any statement that a test/check passes
  without it having been run in this session.
- **[STYLE-04] One file per concept, clear naming.** Flag monolithic files and
  names that need comments to be understood.
- **[STYLE-05] No new dependency without justification.** Flag `go.mod`
  additions with no stated reason; prefer native features first.

## Process rules

- **[PROCESS-01] Findings are specific and actionable.** Flag vague feedback.
  Every finding: where (file:line), why (rule + reasoning), concrete fix.
- **[PROCESS-02] Severity assigned.** Flag findings without a severity
  (blocker/important/suggestion).
- **[PROCESS-03] No unrelated changes.** Flag diffs that touch files outside
  the task's scope without justification.
