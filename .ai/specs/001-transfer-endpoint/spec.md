# Feature Spec: Transfer — domain rules + application use case (logical slice)

Status: Draft (rescoped, clarifications resolved) | Branch: main | Input: "lets create an spec/plan/task to add a /transfer POST http router."

## Decisions confirmed with Rita

- **D-01** The domain **collects uncommitted events**; the application handler drains them and
  publishes each through the publishing port. The domain never holds the publisher.
- **D-02** The **authorization gate is part of this logical slice**: an `Authorizer` port is
  called before any money moves, with a fake in tests. The real gateway adapter is deferred.
- **D-03** **Test fakes are enough** for this slice (plus, at most, a no-op/logging
  implementation). No network call, no broker.
- **D-04** `AccountTransferSucceeded` **gains the `To` account reference**, so every transfer
  event identifies both sides.
- **D-05** A **currency mismatch is refused; no conversion (no FX)**.
- **D-06** Concurrency safety is **deferred with persistence** (recorded known gap, not solved
  here).

## Scope note (rescoped with Rita)

This slice delivers the **logical** transfer flow only: the domain rules, the domain
events, and the application use case that orchestrates them. **No infrastructure**: no
Kafka/broker, no event store, no database writes, no HTTP route, no notification, no real
authorization gateway call. Everything external is expressed as a **port (interface)** with
fakes in tests, so the adapters can land later on top of a settled logical flow.

Follow-up slices (deferred, not this one): `POST /transfer` HTTP route + OpenAPI contract,
persistence of balances and transfers, real authorization adapter, notification + DLQ,
broker publishing / event store.

## Context (facts verified in this session)

- The transfer domain path exists but is unfinished (`internal/domain/account/transfer.go`):
  it loads the payer account, checks `CanTransact()`/`HasBalance()`, loads the payer user and
  checks `CanTransfer()`, then delegates to an **unimplemented** `TransferService`. It never
  loads the payee, never validates the amount, never checks that the two accounts differ, and
  returns untyped `errors.New` values (which cannot be mapped to a client error).
- `Account` has no way to change its balance: there is no debit/credit method, and
  `AccountRepository` only exposes `ByID`, `Add`, `AllAccounts`.
- `Account.HasBalance` uses a strict `>` (account.go:59), so transferring the entire balance
  is wrongly rejected.
- Domain events for transfers already exist as plain structs (`AccountTransferCreated`,
  `AccountTransferSucceeded`, `AccountTransferFailed`) but nothing ever raises them;
  `transfer.go` carries a `// TODO: emit domain events here`.
- `TransferService` and `TransferRepository` are interfaces with no implementation, and no
  `transfers` table exists.
- README defines the endpoint contract as `POST /transfer` with
  `{ "value": 100.0, "payer": 4, "payee": 15 }`; Rita confirmed `from`/`to` are **account
  IDs (UUIDs)** in the domain. The README example's small integers are stale.
- Existing ADRs still hold: ADR-0003 (AP, eventual consistency), ADR-0004/0005 (per-context
  vertical slices), ADR-0006/0007 (invariants in domain factories, plain fields +
  `validateFields`, shared `ConstraintValidationError`).

## User Stories (prioritized, each independently testable)

### US1 — The domain enforces every transfer rule (P1 · MVP)

The transfer rules live in the domain and refuse to produce an invalid or unsafe transfer:
only a COMMON user may pay, the payer account must be able to transact, the amount must be
positive, the two accounts must differ, and the payer's balance must cover the amount.
Balance changes are expressed as behaviour on the account itself, not as arithmetic in a
caller.

Why this priority: rules that live in a caller get duplicated and forgotten; every later
adapter (HTTP, broker, retry) depends on the domain being the single authority.

Independent test: unit-test the domain with plain values (no ports, no fakes) — each rule
rejects, the legal boundary case accepts, and a successful move changes both balances by
exactly the amount.

Acceptance:
- Given a payer account that is able to transact and whose balance is greater than the
  amount, When the transfer is executed, Then the payer balance decreases by exactly the
  amount and the payee balance increases by exactly the amount.
- Given a payer balance exactly equal to the amount, When the transfer is executed, Then it
  is accepted and the payer balance ends at zero.
- Given a payer balance smaller than the amount, When the transfer is executed, Then it is
  rejected with a typed "insufficient balance" error and neither balance changes.
- Given a payer whose account is not ACTIVE, When the transfer is executed, Then it is
  rejected with a typed error and neither balance changes.
- Given a payer who is a SHOPKEEPER, When the transfer is executed, Then it is rejected with
  a typed "not allowed to transfer" error and neither balance changes.
- Given an amount that is zero or negative, When the transfer is executed, Then it is
  rejected with a field-constraint error and neither balance changes.
- Given payer and payee are the same account, When the transfer is executed, Then it is
  rejected with a typed error and no balance changes.
- Given any rejection above, Then the failure is a **typed domain error** (never a bare
  string), so the caller can map it without parsing messages.

### US2 — A transfer attempt is expressed as domain events (P1)

Every attempt produces the transfer facts as domain events, in order, with no broker
involved: the attempt is created, and it either succeeded or failed. A failed attempt
carries why it failed.

Why this priority: events are the project's declared source of truth; the endpoint and the
projections are both derived from them.

Independent test: run one successful and one failed attempt and assert the exact sequence of
events raised, with amounts and account references, and that a failure produces no success
event.

Acceptance:
- Given an attempt that is allowed, When it completes, Then a "transfer created" event is
  raised before a "transfer succeeded" event, both carrying the transfer identity, the amount
  and the account references.
- Given an attempt that is refused, Then "transfer created" is raised before a "transfer
  failed" event carrying the failure reason, and no success event is raised.
- Given an attempt refused by a rule (before any money moves), Then no success event and no
  balance change exist.
- Events are raised for every attempt, including the refused ones, so an audit trail exists
  without a broker.

### US3 — The application use case orchestrates one transfer attempt (P1)

One application command represents "transfer this value from this account to that account".
Its handler is thin orchestration — load what it needs through ports, call the authorization
port before any money moves (D-02), ask the domain to execute the rules, drain the uncommitted
events the domain collected (D-01) and publish them through the publishing port, then return a
result carrying the transfer identity, its final status and the events. All the business
decisions stay in the domain (US1).

Why this priority: the HTTP route and the persistence adapter are both thin wrappers over
this one use case; getting it right once avoids two divergent implementations.

Independent test: with in-memory fakes for every port, submit one valid and one invalid
command and assert the result, the recorded published events, and that no rule logic lives in
the handler (the handler only calls ports and the domain).

Acceptance:
- Given an existing payer account whose owner is a COMMON user and an existing payee account,
  When a valid command is handled, Then the result reports a COMPLETED transfer with its
  identity, both accounts were saved, and the created + succeeded events were published.
- Given a payer or payee that does not exist, When the command is handled, Then the result is
  a typed "not found" error identifying the missing side, nothing is saved, and nothing is
  published.
- Given an identifier that is not a UUID, When the command is handled, Then the result is a
  typed "invalid identifier" error and no port is called.
- Given a refusal from the domain rules, When the command is handled, Then the refusal is
  returned with its type preserved (not flattened to a generic error) and no balance is saved.
- Given the authorization port refuses, When the command is handled, Then the transfer is
  reported FAILED, no balances are saved, and the failure is not reported as success.
- Given the publisher port fails, Then the attempt does not report success.

### US4 — Adapters and contract (P2 — deferred, explicitly out of this slice)

The `POST /transfer` route exposing this use case (contract `{ "value": <number>, "payer":
<account uuid>, "payee": <account uuid> }`), the OpenAPI entry, persistence of balances and
transfers, the real authorization gateway call, and notification with DLQ retry.

Why this priority: every one of them is a thin adapter over US1–US3, and each needs
infrastructure that does not exist in the repository yet (no `transfers` table, no balance
update, no configurable gateway URL, no queue).

Independent test: deferred; tracked as the next slice.

## Edge Cases

- **Exactly-equal balance** must succeed (today the strict `>` rejects it).
- **Zero / negative / non-numeric value** is a field-constraint error.
- **Same account on both sides** is refused.
- **Currency mismatch** between the two accounts: refused, no conversion is performed.
  Decided: refuse, no conversion (D-05).
- **Amount precision** — transferred values stay `float64` for now (ADR-0004 records the
  decimal-money TODO); a logical slice does not make this worse, and it must be revisited
  before real money is stored.
- **Concurrency** — two attempts against the same balance must not both succeed. Real
  protection belongs to the persistence slice; deferred and recorded as a known gap (D-06).
- **Retry / duplicate submission** is not addressed here (no idempotency key); a known
  limitation to record.
- **Partially applied change** — a refused attempt must never leave one balance changed and
  the other not. In this slice both changes happen in memory inside one domain call.
- **Refused-but-recorded**: whether a rule refusal is itself recorded as a FAILED transfer is
  decided in US2 (events are raised for every attempt).

## Functional Requirements

- **FR-001**: The domain MUST refuse a transfer whose payer is not a COMMON user.
- **FR-002**: The domain MUST refuse a transfer whose payer account cannot transact
  (not ACTIVE).
- **FR-003**: The domain MUST refuse a transfer whose amount exceeds the payer's balance, and
  MUST accept an amount equal to it.
- **FR-004**: The domain MUST refuse an amount that is zero or negative, reporting it as the
  offending field.
- **FR-005**: The domain MUST refuse a transfer whose payer and payee are the same account.
- **FR-006**: The domain MUST refuse a transfer between accounts holding different currencies
  and MUST NOT convert between currencies.
- **FR-007**: Every refusal MUST be a typed domain error that preserves the reason, so callers
  map it without parsing messages; validation-of-a-field refusals MUST use the project's
  shared field-constraint error.
- **FR-008**: The domain MUST express a successful transfer as a balance decrease on the payer
  and an equal balance increase on the payee, leaving the total across both accounts
  unchanged.
- **FR-009**: The domain MUST raise a "transfer created" event before the outcome event, and
  exactly one outcome event — succeeded or failed — per attempt, carrying the failure reason
  on failure.
- **FR-010**: The application MUST expose one command + handler for a transfer attempt that
  depends only on ports (accounts, users, transfers, authorization, event publishing) and
  never on a concrete adapter.
- **FR-011**: The application handler MUST NOT contain business rules; every refusal decision
  comes from the domain (FR-001…FR-006).
- **FR-012**: The application handler MUST parse boundary formats (UUID identifiers) and
  report a typed "invalid identifier" error without touching any port.
- **FR-013**: The application handler MUST publish the events raised by the domain through the
  publishing port, and MUST NOT report success when publishing fails.
- **FR-014**: The application result MUST carry the transfer identity, the amount, both
  account references and the terminal status, so a later HTTP adapter can render it without
  re-deriving anything.
- **FR-015**: A refused attempt MUST NOT publish a success event nor persist a changed
  balance.
- **FR-016**: The domain layer MUST stay free of infrastructure concerns (no broker client,
  no HTTP, no database, no serialization, no publisher handle). Events leave the domain as an
  **uncommitted-events collection** on the aggregate, drained by the application handler
  (D-01).
- **FR-017**: `Authorizer` and event `Publisher` MUST be ports; the only implementations in
  this slice are hand-rolled test fakes (plus, at most, a no-op/logging one) — no real network
  or broker call may exist (D-03).
- **FR-018**: `AccountTransferSucceeded` MUST identify both accounts (add the payee reference),
  consistent with the created and failed events (D-04).
- **FR-019**: The authorization gate MUST run before any balance change; an authorization
  refusal MUST leave both balances untouched and MUST NOT publish a success event (D-02).

## Key Entities (no implementation detail)

- **Transfer**: one attempted money movement — identity, payer account, payee account, amount
  (value + currency), status (PENDING → COMPLETED or FAILED) and when it happened. Already
  modelled; gains the behaviour (execution + events).
- **Account**: a wallet — identity, owner, balance, currency, status. Gains the ability to be
  debited and credited under the transfer rules.
- **User**: identity + type (COMMON / SHOPKEEPER) deciding whether the payer may transfer.
  Unchanged.
- **Domain events**: `AccountTransferCreated`, `AccountTransferSucceeded`,
  `AccountTransferFailed` — already defined as plain structs; this slice starts raising them and
  adds the payee reference to `AccountTransferSucceeded` (D-04), so every transfer event
  identifies both sides.

## Success Criteria (measurable)

- **SC-001**: One unit test exists per rule in FR-001…FR-006, each asserting the error type —
  and each fails if the rule is removed.
- **SC-002**: A successful attempt changes both balances by exactly the amount and leaves their
  sum unchanged.
- **SC-003**: Every attempt (allowed or refused) emits exactly one created event and exactly
  one outcome event, in that order, asserted event-by-event.
- **SC-004**: The application handler has no rule branch: its tests exercise the rules solely
  through the domain, and a diff review finds no validation logic in the handler.
- **SC-005**: No file in the domain layer imports an infrastructure concern; the ports used by
  the use case are interfaces, and every test double is a hand-rolled fake.
- **SC-006**: The project's full test suite, build and static analysis (commands in
  `.ai/context/development.md`) pass — run, not asserted — and no existing test regresses.
- **SC-007**: The endpoint contract agreed here (`{value, payer, payee}` with account UUIDs) is
  recorded for the deferred HTTP slice, so it is not re-decided later.

## Assumptions

- `payer`/`payee` are **account** UUIDs (confirmed by Rita); the README's integer example is
  stale and will be corrected when the HTTP slice lands.
- Field names in the eventual contract keep the README's `value`/`payer`/`payee` to avoid
  churn; the values are UUID strings.
- This slice changes no HTTP contract and no database schema; `docs/openapi.yaml` stays as is.
- Money remains `float64` (`MonetaryAmount`, ADR-0004 TODO stands).
- The system has no authentication anywhere, so payer identity is supplied by the caller; the
  use case trusts the command it is given.
- Deposits are out of scope; the deposit stubs stay untouched.
- Authorized-then-moved ordering is kept so no compensation (refund) is needed; the README's
  refund rule applies only to post-move failures such as notification, which is deferred.
- Rule `DDD-07` (no cross-aggregate transaction) is not yet at risk: without persistence there
  is no transaction. It becomes a decision when the persistence slice lands.
