# Plan: Transfer domain rules + application use case (logical slice)

Spec: `specs/001-transfer-endpoint/spec.md` | Branch: main | Phase: plan

## Summary

Deliver the logical transfer flow: the transfer rules enforced inside the domain, the
resulting transfer events collected by the domain as **uncommitted events**, and a thin
application use case that loads the two accounts and the payer user through ports, calls the
authorization port, executes the domain rules, persists the attempt through a single write
seam, and publishes the collected events through a publisher port.

Everything external stays a **port with a hand-rolled fake in tests** (D-01…D-03). No HTTP
route, no Kafka, no event store, no notification, no balance persistence. The two places where
that deliberately leaks are: one new typed not-found error translated by the Postgres adapter
(so the port contract is honest for both fake and real reads), and the verification config
`.pi/gates.json` (so `workflow verify` actually runs the project's checks).

Primary requirement satisfied: FR-001…FR-019 and US1–US3; US4 (adapters) stays deferred.

## Technical Context

- **Language / version**: Go 1.26.0 (module `event-driven-architecture`).
- **Dependencies**: none added. Reuses `github.com/google/uuid`, stdlib `errors`/`time`/`context`.
- **Storage**: untouched — no schema change, no migration, no write path wired. Only the read
  path's not-found error is translated (`sql.ErrNoRows` → typed domain error).
- **HTTP**: untouched. `docs/openapi.yaml` and `router.go` do not change in this slice.
- **Testing**: stdlib `testing` + hand-rolled fakes (no testify), `sqlmock` only for the one
  Postgres read-path test. `httptest` not used (no HTTP in this slice).
- **Target platform**: Linux/macOS, `go build ./cmd/simplified-transfer/`.
- **Project type**: web service (hexagonal DDD + CQRS), only the domain + application layers
  change.
- **Perf goals**: none new; the transfer path must not do I/O inside the domain (aggregate
  state only).
- **Constraints**: domain imports limited to `google/uuid`, `paemuri/brdoc` and pure stdlib
  (DDD-03); no comments in code (STYLE-01); one concept per file (STYLE-04); no unrelated
  changes (PROCESS-03).
- **Scale / scope**: 2 new domain files, 1 rewritten domain file, 3 new application files,
  2 new test files, 1 deleted dead interface, 1 Postgres adapter line + test.

## Goals

- **G-1** Every transfer rule of the spec (FR-001…FR-006) is enforced by the domain and proven
  by a pure domain test that fails if the rule is removed (SC-001).
- **G-2** One attempt produces exactly `created` then exactly one of `succeeded`/`failed`, in
  order, collected by the domain and asserted event-by-event (FR-009, FR-018, SC-003).
- **G-3** The application handler contains no business rule: it parses identifiers, calls
  ports, calls the domain, persists once, publishes the collected events, and maps errors
  (FR-010…FR-015, SC-004).
- **G-4** Authorization runs before any balance change and its refusal changes no balance and
  publishes no success event (FR-019).
- **G-5** Every refusal is a typed error, and business-rule refusals are mapped to application
  sentinels so a future HTTP adapter can map statuses without importing the account context
  (FR-007, DDD-05).
- **G-6** `workflow verify` runs the real project checks (`go build`, `go vet`, `go test`) plus
  static rules, so the implement → review advance is honestly gated (SC-006).

## Non-Goals (explicitly out of this slice)

- `POST /transfer` route, request/response DTOs, OpenAPI entry, error→status mapping (US4).
- Balance persistence, `transfers` table, migration, real transaction/atomicity, concurrency
  safety (D-06) — the write seam is defined but implemented only by a fake.
- Real authorization gateway call, `AUTHORIZER_URL` config, notification, DLQ, Kafka/event
  store, projections, MongoDB.
- Deposits and `deposit.go` / `NewDeposit` / `DepositService` (the existing struct-with-deps
  pattern stays there; migrating it is a separate decision).
- Wiring the use case into `cmd/simplified-transfer/main.go`: none of its ports has a real
  implementation yet, so the running service does not expose any of this.

## Design

### Domain: rules and uncommitted events (`internal/domain/account`)

`Account` owns its own invariants; the transfer factory owns cross-account rules.

```go
type Transfer struct {
	ID     uuid.UUID
	From   AccountID
	To     AccountID
	Amount domain.MonetaryAmount
	Status AccountTransactionStatus
	events []domain.Event
}

func (t *Transfer) UncommittedEvents() []domain.Event
```

Two factories, both keeping the audit trail even when the attempt is refused:

```go
func NewTransfer(id uuid.UUID, from, to *Account, payer *user.User, amount domain.MonetaryAmount) (*Transfer, error)
func NewFailedTransfer(id uuid.UUID, from, to AccountID, amount domain.MonetaryAmount, reason string) *Transfer
```

- `NewTransfer` records `AccountTransferCreated`, then validates (amount positive; payer and
  payee differ; same currency; payer `CanTransfer()`), then `from.Debit(amount)` (account
  ACTIVE + sufficient balance) and `to.Credit(amount)`. On success: `COMPLETED` +
  `AccountTransferSucceeded`. On any refusal: `FAILED` + `AccountTransferFailed{ErrorMessage}`
  and the typed error is returned **together with the transfer**, because the refused attempt
  is the audit record (US2). A non-nil transfer with a non-nil error is intentional here and
  documented in ADR-0008; callers must still check the error.
- `NewFailedTransfer` records `created` + `failed` without touching balances — used for the
  authorization refusal (FR-019).
- `Account.Debit` / `Account.Credit` mutate the in-memory balance; `HasBalance` is fixed to
  `>=` so transferring the whole balance is legal (US1 acceptance).
- `domain.Event` (`EventName() string`) is added in the root package and implemented by the
  three transfer events only, so the publisher port is typed and a future broker has a routing
  key. The remaining events (account/deposit/user) keep their current shape — follow-up item.

### Application: thin orchestration (`internal/application/account/command`)

```go
type TransferMoneyCommand struct {
	PayerAccountID string
	PayeeAccountID string
	Value          float64
}

type TransferMoneyResult struct {
	TransferID     string
	Status         string
	Value          float64
	Currency       string
	PayerAccountID string
	PayeeAccountID string
	Events         []domain.Event
}
```

Handler dependencies (all ports, no concrete adapter):

```go
type TransferMoneyCommandHandler struct {
	accounts   account.AccountRepository
	users      user.UserRepository
	transfers  account.TransferRepository
	authorizer account.TransferAuthorizer
	publisher  TransferEventPublisher
}
```

Order of operations (money never moves before authorization):

1. parse `PayerAccountID` → `ErrInvalidPayerAccountID`; parse `PayeeAccountID` →
   `ErrInvalidPayeeAccountID` (no port touched).
2. `accounts.ByID` payer → `ErrPayerAccountNotFound` (nil account or
   `account.AccountNotFoundError`); same for payee → `ErrPayeeAccountNotFound`.
3. `users.ByID(payer.OwnerId)` → wrap as a persistence failure (the FK guarantees the owner
   exists; a miss is an infrastructure inconsistency → 500 later).
4. `authorizer.Authorize(ctx, transferID)`; on refusal → `NewFailedTransfer`, persist, publish,
   return `ErrTransferNotAuthorized` (no balance moved, no success event).
5. `amount := domain.MonetaryAmount{Currency: from.Balance.Currency, Value: cmd.Value}` (the
   payer's currency is the transfer currency, FR-013).
6. `account.NewTransfer(...)` — the domain executes every rule and collects the events.
7. `transfers.Save(ctx, transfer, from, to)` — exactly once per attempt (also for FAILED), the
   single write seam where the future adapter is atomic (see ADR-0008).
8. publish each `transfer.UncommittedEvents()` → on failure return
   `ErrTransferEventPublishFailed` and never report success (FR-013).
9. if the domain returned a typed error, map it to a sentinel and return it *after* persisting
   and publishing (FR-014/FR-015); field-constraint errors pass through untouched (ADR-0007).

### Ports and their homes

| Port | Home | Why |
|---|---|---|
| `AccountRepository`, `UserRepository` | existing domain ports | unchanged |
| `TransferRepository.Save(ctx, transfer, from, to)` | domain (existing file, method replaced) | one write seam for the attempt + balances; `Add` is replaced because the interface has no implementation |
| `TransferAuthorizer.Authorize(ctx, transferID) error` | `internal/domain/account` | mirrors the `PasswordHasher` precedent: an outbound port kept next to the rule it serves; the domain never calls it |
| `TransferEventPublisher.Publish(ctx, event domain.Event) error` | application command package | outbound port owned by the use case that publishes; the domain only collects events |

### Error chain (DDD-05)

| Domain error | Application sentinel | Future HTTP |
|---|---|---|
| `domain.ConstraintValidationError{amount}` | passed through untouched | 400 |
| `SameAccountTransferError` | `ErrSameAccountTransfer` | 422 |
| `CurrencyMismatchError` | `ErrCurrencyMismatch` | 422 |
| `InsufficientBalanceError` | `ErrInsufficientBalance` | 422 |
| `AccountNotTransactableError` | `ErrAccountNotTransactable` | 409 |
| `TransferNotAllowedError` | `ErrTransferNotAllowed` | 403 |
| `AccountNotFoundError` | `ErrPayerAccountNotFound` / `ErrPayeeAccountNotFound` | 404 |
| authorizer error | `ErrTransferNotAuthorized` | 502/424 |
| publisher error | `ErrTransferEventPublishFailed` | 500 (generic body) |

## Alternatives considered

| Decision | Alternative rejected | Why rejected |
|---|---|---|
| Domain factory returns `(*Transfer, error)` with a non-nil transfer on refusal | Return `nil` on refusal | Loses the audit trail: US2 requires `created` + `failed` events for refused attempts, and there would be no object to collect them from |
| Domain collects uncommitted events; app publishes | Domain holds the publisher port | D-01; would put an infrastructure handle inside the domain and make the domain tests need a fake publisher |
| `TransferRepository.Save(ctx, transfer, from, to)` (single seam) | Add `AccountRepository.Update` | Would break the existing Postgres adapter and the existing fake (both implement that interface), and would give two competing write APIs for one transaction |
| `TransferService` deleted | Keep the interface and implement it later | It holds repositories inside a domain struct and duplicates what the port seam + factory now express; it has no implementation and no caller after the rework |
| `domain.Event` interface with `EventName()` | `[]any` uncommitted events | `any` gives the publisher no routing key and no compile-time safety for a project whose events are the source of truth |
| Authorization in the handler, before the domain executes | Authorization inside the domain factory via a port | The domain must stay free of I/O (DDD-03); the factory would need network stubs in domain tests |
| Authorize first, then validate/execute | Validate first, then authorize (fail-fast) | Accepted trade-off: splitting validation from execution needs a two-phase API (`NewTransfer` + `Complete`) and a second failure path. The gateway is one call per attempt by design, and the mock is cheap. If authorization becomes expensive/rate-limited, split per the `fail-fast-ordering` lesson in `.ai/lessons/` |
| Typed `AccountNotFoundError` + Postgres translation | Handler treats `(nil, nil)` as not-found | The contract would only hold for fakes; production would return a generic 500 instead of 404 in the HTTP slice |

## Constitution Check — GATE

- [x] **Simplicity**: 6 new files, no new dependency, no new abstraction beyond the two ports
  the spec mandates (D-02/D-03) and the single write seam.
- [x] **Test-first**: every task names its test; domain tests are pure values, application tests
  use fakes; each rule test fails if the rule is removed (SC-001).
- [x] **No speculative features**: no idempotency, no retry, no DLQ, no concurrency guard, no
  MongoDB, no projections — all listed as non-goals and recorded as known gaps.
- [x] **Vertical slice**: the slice is independently verifiable by `go test ./...` even though
  it is not wired into the running service (stated as a non-goal).
- [x] **DDD-07 awareness**: exactly one write seam (`TransferRepository.Save`) exists, so the
  future cross-account transaction is a single place to make atomic; nothing in this slice
  mutates two aggregates inside a transaction.

## Complexity Tracking

| Violation | Why needed | Simpler alternative rejected because |
|---|---|---|
| One Postgres adapter change in a "no persistence" slice | The not-found port contract must hold for real reads, not only fakes | Leaving `sql.ErrNoRows` untranslated makes the contract a lie and defers a 404 into a 500 |
| `domain.Event` implemented by only 3 of the 7 events | Keeps the publisher typed without touching unrelated deposit/user/account events (PROCESS-03) | Adding `EventName()` to all 7 events is an unrelated change; follow-up item recorded |

## Structure

```
internal/
├── domain/
│   ├── event.go                                    NEW  Event interface
│   └── account/
│       ├── account.go                              EDIT HasBalance >=; add Debit, Credit
│       ├── account_exceptions.go                   EDIT add AccountNotFoundError
│       ├── account_domain_event.go                 EDIT add To to AccountTransferSucceeded;
│       │                                                EventName() on the 3 transfer events
│       ├── transfer.go                             REWRITE Transfer + NewTransfer factory +
│       │                                                NewFailedTransfer + UncommittedEvents
│       ├── transfer_exceptions.go                  NEW  5 typed transfer errors
│       ├── transfer_authorizer.go                  NEW  TransferAuthorizer port
│       ├── transfer_repository.go                  EDIT Add -> Save(ctx, transfer, from, to)
│       ├── transfer_service.go                     DELETE dead interface after the rework
│       ├── account_test.go                         EDIT Debit/Credit/HasBalance tests
│       └── transfer_test.go                        NEW  rules + events tests
├── application/account/command/
│   ├── errors.go                                   EDIT add transfer sentinels
│   ├── transfer_money_command.go                   NEW  command DTO
│   ├── transfer_event_publisher.go                 NEW  TransferEventPublisher port
│   ├── transfer_money_command_handler.go           NEW  use case + error mapping
│   └── transfer_money_command_handler_test.go      NEW  handler tests with fakes
└── adapters/postgres/
    ├── account_repository.go                       EDIT ByID: ErrNoRows -> AccountNotFoundError
    └── account_repository_test.go                  EDIT not-found test

.pi/gates.json                                      NEW  verify commands + static rules
.ai/decisions/0008-logical-transfer-slice.md         NEW  record the decisions
.ai/architecture/{domain,application,adapters}.md    EDIT layer facts
.ai/context/{development,testing}.md                 EDIT status + test inventory
```

Not touched: `docs/openapi.yaml`, `internal/adapters/http/**`, `cmd/simplified-transfer/main.go`,
`internal/domain/account/{deposit,deposit_service,account_queries,account_repository}.go`,
`internal/domain/user/**`, `internal/application/user/**`.

## Test plan (one test per rule and per acceptance criterion)

Domain — `internal/domain/account/transfer_test.go` (pure values, no fakes):

| Test | Proves |
|---|---|
| `TestNewTransfer_Success` | balances moved by exactly the amount; sum unchanged; status COMPLETED |
| `TestNewTransfer_ExactBalance` | whole balance transfer accepted; payer ends at zero |
| `TestNewTransfer_InsufficientBalance` | typed `InsufficientBalanceError`; balances untouched |
| `TestNewTransfer_InactiveAccount` | typed `AccountNotTransactableError`; balances untouched |
| `TestNewTransfer_ShopkeeperPayer` | typed `TransferNotAllowedError`; balances untouched |
| `TestNewTransfer_NonPositiveAmount` | `ConstraintValidationError{Field:"amount"}` |
| `TestNewTransfer_SameAccount` | typed `SameAccountTransferError` |
| `TestNewTransfer_CurrencyMismatch` | typed `CurrencyMismatchError` |
| `TestNewTransfer_Events` | exactly created → succeeded on success |
| `TestNewTransfer_FailureEvents` | created → failed with reason, no succeeded event |
| `TestNewFailedTransfer_Events` | created → failed, no balance change |
| `TestAccount_Debit` / `TestAccount_Credit` | exact arithmetic; `HasBalance` boundary `>=` |

Application — `internal/application/account/command/transfer_money_command_handler_test.go`
(fakes: existing `fakeAccountRepository` extended for not-found, new `fakeUserRepository`,
`fakeTransferRepository` recording saves, `stubAuthorizer`, `recordingPublisher`):

| Test | Proves |
|---|---|
| `TestTransferMoneyCommandHandler_Success` | result fields, one save, events published in order |
| `TestTransferMoneyCommandHandler_InvalidPayerID` / `..._InvalidPayeeID` | sentinel, no port called |
| `TestTransferMoneyCommandHandler_PayerAccountNotFound` / `..._PayeeAccountNotFound` | sentinels identify the missing side |
| `TestTransferMoneyCommandHandler_NotAuthorized` | FAILED recorded, no balance change, no success event |
| `TestTransferMoneyCommandHandler_PublisherFailure` | error, not reported as success |
| `TestTransferMoneyCommandHandler_DomainRefusal` | sentinel preserved; attempt persisted; failure event published |
| `TestTransferMoneyCommandHandler_SaveFailure` | persistence failure surfaces as an error |

Postgres — `internal/adapters/postgres/account_repository_test.go`:
`TestAccountRepository_ByIDNotFound` (`sqlmock` rows with no result → `AccountNotFoundError`).

## Verification config (`.pi/gates.json`)

```json
{
  "verify": {
    "commands": ["go build ./cmd/simplified-transfer/", "go vet ./...", "go test ./..."]
  },
  "rules": [
    { "id": "domain-purity", "description": "domain imports no infrastructure", "mode": "forbidden", "command": "grep -rn 'net/http\\|gin-gonic\\|database/sql' internal/domain/ 2>/dev/null" },
    { "id": "gofmt", "description": "all Go files formatted", "mode": "forbidden", "command": "gofmt -l internal cmd 2>/dev/null" }
  ]
}
```

## Open Questions

None blocking. Deferred and recorded: concurrency safety (D-06), idempotency, notification +
DLQ, broker/event-store publishing, balance persistence and the atomicity ADR, `EventName()`
for the remaining events, migration of the `NewDeposit`/`DepositService` pattern.

## Snippets for the task phase (verbatim from current files)

`internal/domain/account/account.go:59` — the bug to fix:

```go
func (a Account) HasBalance(amount domain.MonetaryAmount) bool {
	return a.Balance.Value > amount.Value
}
```

`internal/domain/account/transfer_repository.go` — the method to replace:

```go
type TransferRepository interface {
	ByID(ctx context.Context, id string) (*Transfer, error)
	Add(ctx context.Context, transfer *Transfer) error

	AllTransfers(ctx context.Context) ([]*Transfer, error)
}
```

`internal/domain/account/transfer.go` — replaced in full (current body):

```go
type NewTransfer struct {
	userRepository    user.UserRepository
	transferService   TransferService
	accountRepository AccountRepository
}

func (t NewTransfer) CreateTransfer(ctx context.Context, from, to AccountID, amount domain.MonetaryAmount) (*Transfer, error) {
	fromAccount, err := t.accountRepository.ByID(ctx, from)
	...
}
```

`internal/domain/account/account_domain_event.go:24` — the event gaining `To`:

```go
type AccountTransferSucceeded struct {
	ID        uuid.UUID
	CreatedAt time.Time
	Amount    float64
	From      AccountID
}
```

`internal/application/account/command/errors.go` — the file gaining sentinels:

```go
var (
	ErrAccountAlreadyExists = errors.New("account already exists for this owner")
	ErrOwnerNotFound        = errors.New("owner not found")
	ErrInvalidOwnerID       = errors.New("owner id must be a valid uuid")
)
```

`internal/adapters/postgres/account_repository.go` — the read path gaining translation:

```go
func (r *AccountRepository) ByID(ctx context.Context, id account.AccountID) (*account.Account, error) {
	row := r.db.QueryRowContext(ctx, selectAccountByIDSQL, uuid.UUID(id))
	acc, err := scanAccount(row)
	if err != nil {
		return nil, err
	}
	return acc, nil
}
```

`internal/application/account/command/create_account_command_handler_test.go:31` — the fake to
extend for not-found (currently returns a nil account without error):

```go
func (f *fakeAccountRepository) ByID(_ context.Context, id account.AccountID) (*account.Account, error) {
	return f.accounts[uuid.UUID(id)], nil
}
```
