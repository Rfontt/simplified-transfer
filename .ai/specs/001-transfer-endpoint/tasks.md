# Tasks: Transfer domain rules + application use case (logical slice)

Spec: `specs/001-transfer-endpoint/spec.md` | Plan: `specs/001-transfer-endpoint/plan.md`

Apply these tasks strictly in order. One task = one file = one focused change. Verify each
`Done` before moving on. Source names/labels in the `NEW` blocks are already reviewed — do not
rename or restructure them.

## T001 — Add the project verification config

File: `.pi/gates.json`
Test: not applicable (configuration) — verified by `workflow action=verify` running the commands
Done: `.pi/gates.json` exists and `go build ./cmd/simplified-transfer/`, `go vet ./...`, `go test ./...` all exit 0 when run from the repo root

OLD:
```
(file does not exist)
```

NEW:
```json
{
  "verify": {
    "commands": ["go build ./cmd/simplified-transfer/", "go vet ./...", "go test ./..."]
  },
  "rules": [
    {
      "id": "domain-purity",
      "description": "domain imports no infrastructure",
      "mode": "forbidden",
      "command": "grep -rn 'net/http\\|gin-gonic\\|database/sql' internal/domain/ 2>/dev/null"
    },
    {
      "id": "gofmt",
      "description": "all Go files formatted",
      "mode": "forbidden",
      "command": "gofmt -l internal cmd 2>/dev/null"
    }
  ]
}
```

## T002 — Add the domain Event interface

File: `internal/domain/event.go`
Test: not applicable (tiny interface) — covered by `TestNewTransfer_Events` and `TestNewTransfer_FailureEvents`
Done: `go build ./...` succeeds and the file has no comments

OLD:
```
(file does not exist)
```

NEW:
```go
package domain

type Event interface {
	EventName() string
}
```

## T003 — Add the typed transfer errors

File: `internal/domain/account/transfer_exceptions.go`
Test: `internal/domain/account/transfer_test.go` (T012) asserts each of these types via `errors.As`
Done: `go build ./...` succeeds; no `errors.New` string errors remain in the transfer path

OLD:
```
(file does not exist)
```

NEW:
```go
package account

import (
	"fmt"

	"github.com/google/uuid"
)

type SameAccountTransferError struct {
	AccountID AccountID
}

func (e *SameAccountTransferError) Error() string {
	return fmt.Sprintf("account %s cannot transfer to itself", uuid.UUID(e.AccountID))
}

type CurrencyMismatchError struct {
	FromCurrency string
	ToCurrency   string
}

func (e *CurrencyMismatchError) Error() string {
	return fmt.Sprintf("cannot transfer between accounts with different currencies: %s and %s", e.FromCurrency, e.ToCurrency)
}

type InsufficientBalanceError struct {
	AccountID AccountID
	Balance   float64
	Amount    float64
}

func (e *InsufficientBalanceError) Error() string {
	return fmt.Sprintf("account %s has insufficient balance: %v < %v", uuid.UUID(e.AccountID), e.Balance, e.Amount)
}

type AccountNotTransactableError struct {
	AccountID AccountID
	Status    AccountStatus
}

func (e *AccountNotTransactableError) Error() string {
	return fmt.Sprintf("account %s with status %s is not allowed to transact", uuid.UUID(e.AccountID), e.Status)
}

type TransferNotAllowedError struct {
	UserType string
}

func (e *TransferNotAllowedError) Error() string {
	return fmt.Sprintf("%s users are not allowed to transfer", e.UserType)
}
```

## T004 — Add the account not-found error

File: `internal/domain/account/account_exceptions.go`
Test: `internal/adapters/postgres/account_repository_test.go` (T018) asserts the type
Done: `go build ./...` succeeds and `AccountNotFoundError` is declared with an `AccountID string` field

OLD:
```go
type OwnerNotFoundError struct {
	OwnerID string
}

func (e *OwnerNotFoundError) Error() string {
	return fmt.Sprintf("owner %s not found", e.OwnerID)
}
```

NEW:
```go
type OwnerNotFoundError struct {
	OwnerID string
}

func (e *OwnerNotFoundError) Error() string {
	return fmt.Sprintf("owner %s not found", e.OwnerID)
}

type AccountNotFoundError struct {
	AccountID string
}

func (e *AccountNotFoundError) Error() string {
	return fmt.Sprintf("account %s not found", e.AccountID)
}
```

## T005 — Make the balance rules correct and let the account move its own balance

File: `internal/domain/account/account.go`
Test: `internal/domain/account/account_test.go` — `TestAccount_HasBalanceExactAmount`, `TestAccount_DebitAndCredit`, `TestAccount_DebitNotTransactable`, `TestAccount_DebitInsufficientBalance` (T011)
Done: `go test ./internal/domain/account/...` green; `HasBalance` compares with `>=`; `Debit` refuses an inactive account and an insufficient balance without mutating; `Credit` adds and returns nothing

OLD:
```go
func (a Account) CanTransact() bool {
	return a.Status == ACTIVE
}

func (a Account) HasBalance(amount domain.MonetaryAmount) bool {
	return a.Balance.Value > amount.Value
}
```

NEW:
```go
func (a Account) CanTransact() bool {
	return a.Status == ACTIVE
}

func (a Account) HasBalance(amount domain.MonetaryAmount) bool {
	return a.Balance.Value >= amount.Value
}

func (a *Account) Debit(amount domain.MonetaryAmount) error {
	if !a.CanTransact() {
		return &AccountNotTransactableError{AccountID: a.ID, Status: a.Status}
	}
	if !a.HasBalance(amount) {
		return &InsufficientBalanceError{AccountID: a.ID, Balance: a.Balance.Value, Amount: amount.Value}
	}
	a.Balance.Value -= amount.Value
	return nil
}

func (a *Account) Credit(amount domain.MonetaryAmount) {
	a.Balance.Value += amount.Value
}
```

## T006 — Transfer events identify both sides and carry a name

File: `internal/domain/account/account_domain_event.go`
Test: `internal/domain/account/transfer_test.go` — `TestNewTransfer_Events`, `TestNewTransfer_FailureEvents` (T012)
Done: `AccountTransferSucceeded` has a `To AccountID` field; each of the three transfer events has an `EventName()` method returning its own type name; `go build ./...` succeeds

OLD:
```go
type AccountTransferSucceeded struct {
	ID        uuid.UUID
	CreatedAt time.Time
	Amount    float64
	From      AccountID
}
```

NEW:
```go
type AccountTransferSucceeded struct {
	ID        uuid.UUID
	CreatedAt time.Time
	Amount    float64
	From      AccountID
	To        AccountID
}

func (AccountTransferCreated) EventName() string {
	return "AccountTransferCreated"
}

func (AccountTransferSucceeded) EventName() string {
	return "AccountTransferSucceeded"
}

func (AccountTransferFailed) EventName() string {
	return "AccountTransferFailed"
}
```

## T007 — Turn the transfer repository into the single write seam

File: `internal/domain/account/transfer_repository.go`
Test: `internal/application/account/command/transfer_money_command_handler_test.go` — `TestTransferMoneyCommandHandler_SaveFailure` (T020)
Done: `go build ./...` succeeds; the interface exposes `Save(ctx, transfer, from, to)` and no longer `Add`; no implementation of the old method remains anywhere (`grep -rn "func .*Add(ctx context.Context, transfer" internal/` finds nothing)

OLD:
```go
package account

import (
	"context"
)

type TransferRepository interface {
	ByID(ctx context.Context, id string) (*Transfer, error)
	Add(ctx context.Context, transfer *Transfer) error

	AllTransfers(ctx context.Context) ([]*Transfer, error)
}
```

NEW:
```go
package account

import "context"

type TransferRepository interface {
	ByID(ctx context.Context, id string) (*Transfer, error)
	Save(ctx context.Context, transfer *Transfer, from, to *Account) error

	AllTransfers(ctx context.Context) ([]*Transfer, error)
}
```

## T008 — Add the transfer authorization port

File: `internal/domain/account/transfer_authorizer.go`
Test: `internal/application/account/command/transfer_money_command_handler_test.go` — `TestTransferMoneyCommandHandler_NotAuthorized` (T020)
Done: `go build ./...` succeeds; the port is an interface with a single `Authorize(ctx, transferID) error` method and no implementation exists in the repo

OLD:
```
(file does not exist)
```

NEW:
```go
package account

import "context"

type TransferAuthorizer interface {
	Authorize(ctx context.Context, transferID string) error
}
```

## T009 — Replace the transfer flow with the domain factory that owns the rules and collects events

File: `internal/domain/account/transfer.go`
Test: `internal/domain/account/transfer_test.go` (T012) — all transfer tests
Done: `go test ./internal/domain/account/...` green; `NewTransfer`/`NewFailedTransfer`/`UncommittedEvents` exist; the `NewTransfer` struct, its `CreateTransfer` method, every untyped `errors.New` and the `// TODO: emit domain events here` are gone; the file imports only `domain`, `user`, `time` and `uuid`

OLD:
```go
package account

import (
	"context"
	"errors"
	"event-driven-architecture/internal/domain"
	"event-driven-architecture/internal/domain/user"
	"fmt"

	"github.com/google/uuid"
)

type Transfer struct {
	ID     uuid.UUID
	From   AccountID
	To     AccountID
	Amount domain.MonetaryAmount
	Status AccountTransactionStatus
}

type NewTransfer struct {
	userRepository    user.UserRepository
	transferService   TransferService
	accountRepository AccountRepository
}

func (t NewTransfer) CreateTransfer(ctx context.Context, from, to AccountID, amount domain.MonetaryAmount) (*Transfer, error) {
	fromAccount, err := t.accountRepository.ByID(ctx, from)
	if err != nil {
		return nil, err
	}

	if !fromAccount.CanTransact() || !fromAccount.HasBalance(amount) {
		return nil, errors.New("sender account is not allowed to perform transfers")
	}

	userFrom, err := t.userRepository.ByID(ctx, fromAccount.OwnerId)
	if err != nil {
		return nil, err
	}

	if !userFrom.CanTransfer() {
		return nil, fmt.Errorf("%v user type is not allowed to transfer", userFrom.Type)
	}

	result, err := t.transferService.TransferFunds(from, to, amount)
	if err != nil {
		return nil, err
	}

	// TODO: emit domain events here

	return result, nil
}
```

NEW:
```go
package account

import (
	"time"

	"event-driven-architecture/internal/domain"
	"event-driven-architecture/internal/domain/user"

	"github.com/google/uuid"
)

type Transfer struct {
	ID     uuid.UUID
	From   AccountID
	To     AccountID
	Amount domain.MonetaryAmount
	Status AccountTransactionStatus

	events []domain.Event
}

func NewTransfer(id uuid.UUID, from, to *Account, payer *user.User, amount domain.MonetaryAmount) (*Transfer, error) {
	transfer := newPendingTransfer(id, from.ID, to.ID, amount)
	if err := transfer.validate(from, to, payer); err != nil {
		return transfer.failed(err.Error()), err
	}
	if err := from.Debit(amount); err != nil {
		return transfer.failed(err.Error()), err
	}
	to.Credit(amount)
	transfer.Status = COMPLETED
	transfer.record(AccountTransferSucceeded{
		ID:        id,
		CreatedAt: time.Now(),
		Amount:    amount.Value,
		From:      from.ID,
		To:        to.ID,
	})
	return transfer, nil
}

func NewFailedTransfer(id uuid.UUID, from, to AccountID, amount domain.MonetaryAmount, reason string) *Transfer {
	return newPendingTransfer(id, from, to, amount).failed(reason)
}

func (t *Transfer) UncommittedEvents() []domain.Event {
	return t.events
}

func (t *Transfer) validate(from, to *Account, payer *user.User) error {
	if !t.Amount.IsPositive() {
		return &domain.ConstraintValidationError{Field: "amount"}
	}
	if t.From == t.To {
		return &SameAccountTransferError{AccountID: t.From}
	}
	if from.Balance.Currency != to.Balance.Currency {
		return &CurrencyMismatchError{FromCurrency: from.Balance.Currency, ToCurrency: to.Balance.Currency}
	}
	if !payer.CanTransfer() {
		return &TransferNotAllowedError{UserType: payer.Type}
	}
	return nil
}

func (t *Transfer) failed(reason string) *Transfer {
	t.Status = FAILED
	t.record(AccountTransferFailed{
		ID:           t.ID,
		CreatedAt:    time.Now(),
		Amount:       t.Amount.Value,
		To:           t.To,
		From:         t.From,
		ErrorMessage: reason,
	})
	return t
}

func (t *Transfer) record(event domain.Event) {
	t.events = append(t.events, event)
}

func newPendingTransfer(id uuid.UUID, from, to AccountID, amount domain.MonetaryAmount) *Transfer {
	transfer := &Transfer{ID: id, From: from, To: to, Amount: amount, Status: PENDING}
	transfer.record(AccountTransferCreated{
		ID:        id,
		CreatedAt: time.Now(),
		Amount:    amount.Value,
		To:        to,
		From:      from,
	})
	return transfer
}
```

## T010 — Delete the dead transfer service interface

File: `internal/domain/account/transfer_service.go`
Test: not applicable (deletion) — covered by `go build ./...` and `go vet ./...`
Done: `rm internal/domain/account/transfer_service.go` executed; `grep -rn "TransferService" internal/ cmd/` returns nothing; `go build ./...` succeeds

OLD:
```go
package account

import "event-driven-architecture/internal/domain"

type TransferService interface {
	TransferFunds(from, to AccountID, amount domain.MonetaryAmount) (*Transfer, error)
}
```

NEW:
```
(file removed)
```

## T011 — Test the account balance rules

File: `internal/domain/account/account_test.go`
Test: `go test ./internal/domain/account/ -run "TestAccount_" -v`
Done: the four tests pass; `TestAccount_HasBalanceExactAmount` fails if `HasBalance` is reverted to `>`; the two debit error tests assert the balance is unchanged

OLD:
```go
func TestNewAccount_TrimsCurrency(t *testing.T) {
	acc, err := NewAccount(AccountID(uuid.New()), user.ID(uuid.New()), domain.MonetaryAmount{Currency: " BRL ", Value: 100})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if acc.Balance.Currency != "BRL" {
		t.Errorf("expected trimmed currency, got %q", acc.Balance.Currency)
	}
}
```

NEW:
```go
func TestNewAccount_TrimsCurrency(t *testing.T) {
	acc, err := NewAccount(AccountID(uuid.New()), user.ID(uuid.New()), domain.MonetaryAmount{Currency: " BRL ", Value: 100})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if acc.Balance.Currency != "BRL" {
		t.Errorf("expected trimmed currency, got %q", acc.Balance.Currency)
	}
}

func newTestAccount(t *testing.T, ownerID user.ID, currency string, balance float64) *Account {
	t.Helper()
	acc, err := NewAccount(AccountID(uuid.New()), ownerID, domain.MonetaryAmount{Currency: currency, Value: balance})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return acc
}

func TestAccount_HasBalanceExactAmount(t *testing.T) {
	acc := newTestAccount(t, user.ID(uuid.New()), "BRL", 100)

	if !acc.HasBalance(domain.MonetaryAmount{Currency: "BRL", Value: 100}) {
		t.Error("expected an amount equal to the balance to be sufficient")
	}
	if acc.HasBalance(domain.MonetaryAmount{Currency: "BRL", Value: 100.01}) {
		t.Error("expected an amount above the balance to be insufficient")
	}
}

func TestAccount_DebitAndCredit(t *testing.T) {
	acc := newTestAccount(t, user.ID(uuid.New()), "BRL", 100)

	if err := acc.Debit(domain.MonetaryAmount{Currency: "BRL", Value: 30}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if acc.Balance.Value != 70 {
		t.Errorf("expected balance 70, got %v", acc.Balance.Value)
	}

	acc.Credit(domain.MonetaryAmount{Currency: "BRL", Value: 10})
	if acc.Balance.Value != 80 {
		t.Errorf("expected balance 80, got %v", acc.Balance.Value)
	}
}

func TestAccount_DebitNotTransactable(t *testing.T) {
	acc := newTestAccount(t, user.ID(uuid.New()), "BRL", 100)
	acc.Status = BLOCKED

	err := acc.Debit(domain.MonetaryAmount{Currency: "BRL", Value: 10})
	var notTransactable *AccountNotTransactableError
	if !errors.As(err, &notTransactable) {
		t.Fatalf("expected AccountNotTransactableError, got %v", err)
	}
	if acc.Balance.Value != 100 {
		t.Errorf("expected the balance to be unchanged, got %v", acc.Balance.Value)
	}
}

func TestAccount_DebitInsufficientBalance(t *testing.T) {
	acc := newTestAccount(t, user.ID(uuid.New()), "BRL", 100)

	err := acc.Debit(domain.MonetaryAmount{Currency: "BRL", Value: 101})
	var insufficient *InsufficientBalanceError
	if !errors.As(err, &insufficient) {
		t.Fatalf("expected InsufficientBalanceError, got %v", err)
	}
	if acc.Balance.Value != 100 {
		t.Errorf("expected the balance to be unchanged, got %v", acc.Balance.Value)
	}
}
```

## T012 — Test the transfer rules and the collected events

File: `internal/domain/account/transfer_test.go`
Test: `go test ./internal/domain/account/ -run "TestNewTransfer|TestNewFailedTransfer" -v`
Done: all tests pass; each rule test removes its rule and fails; the success test asserts both balances and their unchanged sum; the failure tests assert exactly `AccountTransferCreated` + `AccountTransferFailed`

OLD:
```
(file does not exist)
```

NEW:
```go
package account

import (
	"errors"
	"testing"

	"event-driven-architecture/internal/domain"
	"event-driven-architecture/internal/domain/user"

	"github.com/google/uuid"
)

type stubHasher struct{}

func (stubHasher) Hash(plain string) (string, error) {
	return "hashed:" + plain, nil
}

func newTestPayer(t *testing.T, id user.ID, userType string) *user.User {
	t.Helper()
	payer, err := user.NewUser(id, stubHasher{}, "Rita Fontenele", "52998224725", "rita@example.com", "secret", userType)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return payer
}

func transferFixture(t *testing.T, fromBalance, toBalance float64) (*Account, *Account, *user.User) {
	t.Helper()
	ownerID := user.ID(uuid.New())
	from := newTestAccount(t, ownerID, "BRL", fromBalance)
	to := newTestAccount(t, user.ID(uuid.New()), "BRL", toBalance)
	return from, to, newTestPayer(t, ownerID, user.COMMON)
}

func assertTransferEvents(t *testing.T, transfer *Transfer, names ...string) {
	t.Helper()
	events := transfer.UncommittedEvents()
	if len(events) != len(names) {
		t.Fatalf("expected %d events, got %d", len(names), len(events))
	}
	for i, name := range names {
		if events[i].EventName() != name {
			t.Errorf("event %d: expected %s, got %s", i, name, events[i].EventName())
		}
	}
}

func TestNewTransfer_Success(t *testing.T) {
	from, to, payer := transferFixture(t, 100, 0)

	transfer, err := NewTransfer(uuid.New(), from, to, payer, domain.MonetaryAmount{Currency: "BRL", Value: 40})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if transfer.Status != COMPLETED {
		t.Errorf("expected COMPLETED, got %s", transfer.Status)
	}
	if from.Balance.Value != 60 {
		t.Errorf("expected payer balance 60, got %v", from.Balance.Value)
	}
	if to.Balance.Value != 40 {
		t.Errorf("expected payee balance 40, got %v", to.Balance.Value)
	}
	if from.Balance.Value+to.Balance.Value != 100 {
		t.Errorf("expected the total to stay 100, got %v", from.Balance.Value+to.Balance.Value)
	}
	assertTransferEvents(t, transfer, "AccountTransferCreated", "AccountTransferSucceeded")
}

func TestNewTransfer_ExactBalance(t *testing.T) {
	from, to, payer := transferFixture(t, 100, 0)

	transfer, err := NewTransfer(uuid.New(), from, to, payer, domain.MonetaryAmount{Currency: "BRL", Value: 100})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if from.Balance.Value != 0 {
		t.Errorf("expected payer balance 0, got %v", from.Balance.Value)
	}
	if to.Balance.Value != 100 {
		t.Errorf("expected payee balance 100, got %v", to.Balance.Value)
	}
	if transfer.Status != COMPLETED {
		t.Errorf("expected COMPLETED, got %s", transfer.Status)
	}
}

func TestNewTransfer_InsufficientBalance(t *testing.T) {
	from, to, payer := transferFixture(t, 100, 0)

	transfer, err := NewTransfer(uuid.New(), from, to, payer, domain.MonetaryAmount{Currency: "BRL", Value: 100.01})
	var insufficient *InsufficientBalanceError
	if !errors.As(err, &insufficient) {
		t.Fatalf("expected InsufficientBalanceError, got %v", err)
	}
	if from.Balance.Value != 100 || to.Balance.Value != 0 {
		t.Errorf("expected the balances to be unchanged, got %v and %v", from.Balance.Value, to.Balance.Value)
	}
	if transfer.Status != FAILED {
		t.Errorf("expected FAILED, got %s", transfer.Status)
	}
	assertTransferEvents(t, transfer, "AccountTransferCreated", "AccountTransferFailed")
}

func TestNewTransfer_NotTransactableAccount(t *testing.T) {
	from, to, payer := transferFixture(t, 100, 0)
	from.Status = BLOCKED

	transfer, err := NewTransfer(uuid.New(), from, to, payer, domain.MonetaryAmount{Currency: "BRL", Value: 10})
	var notTransactable *AccountNotTransactableError
	if !errors.As(err, &notTransactable) {
		t.Fatalf("expected AccountNotTransactableError, got %v", err)
	}
	if from.Balance.Value != 100 || to.Balance.Value != 0 {
		t.Errorf("expected the balances to be unchanged, got %v and %v", from.Balance.Value, to.Balance.Value)
	}
	assertTransferEvents(t, transfer, "AccountTransferCreated", "AccountTransferFailed")
}

func TestNewTransfer_ShopkeeperPayer(t *testing.T) {
	from, to, _ := transferFixture(t, 100, 0)
	payer := newTestPayer(t, from.OwnerId, user.SHOPKEEPER)

	transfer, err := NewTransfer(uuid.New(), from, to, payer, domain.MonetaryAmount{Currency: "BRL", Value: 10})
	var notAllowed *TransferNotAllowedError
	if !errors.As(err, &notAllowed) {
		t.Fatalf("expected TransferNotAllowedError, got %v", err)
	}
	if from.Balance.Value != 100 || to.Balance.Value != 0 {
		t.Errorf("expected the balances to be unchanged, got %v and %v", from.Balance.Value, to.Balance.Value)
	}
	assertTransferEvents(t, transfer, "AccountTransferCreated", "AccountTransferFailed")
}

func TestNewTransfer_NonPositiveAmount(t *testing.T) {
	from, to, payer := transferFixture(t, 100, 0)

	for _, value := range []float64{0, -1} {
		_, err := NewTransfer(uuid.New(), from, to, payer, domain.MonetaryAmount{Currency: "BRL", Value: value})
		assertConstraintError(t, err, "amount")
	}
}

func TestNewTransfer_SameAccount(t *testing.T) {
	from, _, payer := transferFixture(t, 100, 0)

	transfer, err := NewTransfer(uuid.New(), from, from, payer, domain.MonetaryAmount{Currency: "BRL", Value: 10})
	var sameAccount *SameAccountTransferError
	if !errors.As(err, &sameAccount) {
		t.Fatalf("expected SameAccountTransferError, got %v", err)
	}
	if from.Balance.Value != 100 {
		t.Errorf("expected the balance to be unchanged, got %v", from.Balance.Value)
	}
	assertTransferEvents(t, transfer, "AccountTransferCreated", "AccountTransferFailed")
}

func TestNewTransfer_CurrencyMismatch(t *testing.T) {
	ownerID := user.ID(uuid.New())
	from := newTestAccount(t, ownerID, "BRL", 100)
	to := newTestAccount(t, user.ID(uuid.New()), "USD", 0)
	payer := newTestPayer(t, ownerID, user.COMMON)

	transfer, err := NewTransfer(uuid.New(), from, to, payer, domain.MonetaryAmount{Currency: "BRL", Value: 10})
	var mismatch *CurrencyMismatchError
	if !errors.As(err, &mismatch) {
		t.Fatalf("expected CurrencyMismatchError, got %v", err)
	}
	if from.Balance.Value != 100 || to.Balance.Value != 0 {
		t.Errorf("expected the balances to be unchanged, got %v and %v", from.Balance.Value, to.Balance.Value)
	}
	assertTransferEvents(t, transfer, "AccountTransferCreated", "AccountTransferFailed")
}

func TestNewFailedTransfer(t *testing.T) {
	from, to, _ := transferFixture(t, 100, 0)

	transfer := NewFailedTransfer(uuid.New(), from.ID, to.ID, domain.MonetaryAmount{Currency: "BRL", Value: 10}, "gateway unavailable")
	if transfer.Status != FAILED {
		t.Errorf("expected FAILED, got %s", transfer.Status)
	}
	if from.Balance.Value != 100 || to.Balance.Value != 0 {
		t.Errorf("expected the balances to be unchanged, got %v and %v", from.Balance.Value, to.Balance.Value)
	}
	assertTransferEvents(t, transfer, "AccountTransferCreated", "AccountTransferFailed")
}
```

## T013 — Add the transfer application sentinels

File: `internal/application/account/command/errors.go`
Test: `internal/application/account/command/transfer_money_command_handler_test.go` (T020) asserts each sentinel with `errors.Is`
Done: `go build ./...` succeeds; the file declares the eleven transfer sentinels; existing sentinels are unchanged

OLD:
```go
var (
	ErrAccountAlreadyExists = errors.New("account already exists for this owner")
	ErrOwnerNotFound        = errors.New("owner not found")
	ErrInvalidOwnerID       = errors.New("owner id must be a valid uuid")
)
```

NEW:
```go
var (
	ErrAccountAlreadyExists = errors.New("account already exists for this owner")
	ErrOwnerNotFound        = errors.New("owner not found")
	ErrInvalidOwnerID       = errors.New("owner id must be a valid uuid")

	ErrInvalidPayerAccountID      = errors.New("payer account id must be a valid uuid")
	ErrInvalidPayeeAccountID      = errors.New("payee account id must be a valid uuid")
	ErrPayerAccountNotFound       = errors.New("payer account not found")
	ErrPayeeAccountNotFound       = errors.New("payee account not found")
	ErrTransferNotAuthorized      = errors.New("transfer not authorized")
	ErrInsufficientBalance        = errors.New("insufficient balance")
	ErrAccountNotTransactable     = errors.New("payer account is not allowed to transact")
	ErrTransferNotAllowed         = errors.New("payer is not allowed to transfer")
	ErrSameAccountTransfer        = errors.New("payer and payee accounts must differ")
	ErrCurrencyMismatch           = errors.New("accounts must hold the same currency")
	ErrTransferEventPublishFailed = errors.New("failed to publish transfer events")
)
```

## T014 — Add the transfer command DTO

File: `internal/application/account/command/transfer_money_command.go`
Test: `internal/application/account/command/transfer_money_command_handler_test.go` — `TestTransferMoneyCommandHandler_Success` (T020)
Done: `go build ./...` succeeds; the struct has exactly the fields `PayerAccountID`, `PayeeAccountID`, `Value` and no methods

OLD:
```
(file does not exist)
```

NEW:
```go
package command

type TransferMoneyCommand struct {
	PayerAccountID string
	PayeeAccountID string
	Value          float64
}
```

## T015 — Add the transfer event publisher port

File: `internal/application/account/command/transfer_event_publisher.go`
Test: `internal/application/account/command/transfer_money_command_handler_test.go` — `TestTransferMoneyCommandHandler_PublisherFailure` (T020)
Done: `go build ./...` succeeds; the port takes `(ctx, event domain.Event) error`; the only implementation in the repo is the test fake

OLD:
```
(file does not exist)
```

NEW:
```go
package command

import (
	"context"

	"event-driven-architecture/internal/domain"
)

type TransferEventPublisher interface {
	Publish(ctx context.Context, event domain.Event) error
}
```

## T016 — Add the thin transfer use case

File: `internal/application/account/command/transfer_money_command_handler.go`
Test: `internal/application/account/command/transfer_money_command_handler_test.go` (T020) — all handler tests
Done: `go test ./internal/application/account/command/...` green; the handler holds no rule branch (no currency, balance, status or user-type comparison); the order is parse → load → authorize → domain → save → publish → map error

OLD:
```
(file does not exist)
```

NEW:
```go
package command

import (
	"context"
	"errors"
	"fmt"

	"event-driven-architecture/internal/domain"
	"event-driven-architecture/internal/domain/account"
	"event-driven-architecture/internal/domain/user"

	"github.com/google/uuid"
)

type TransferMoneyUseCase interface {
	Handle(ctx context.Context, cmd TransferMoneyCommand) (*TransferMoneyResult, error)
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

type TransferMoneyCommandHandler struct {
	accounts   account.AccountRepository
	users      user.UserRepository
	transfers  account.TransferRepository
	authorizer account.TransferAuthorizer
	publisher  TransferEventPublisher
}

func NewTransferMoneyCommandHandler(
	accounts account.AccountRepository,
	users user.UserRepository,
	transfers account.TransferRepository,
	authorizer account.TransferAuthorizer,
	publisher TransferEventPublisher,
) *TransferMoneyCommandHandler {
	return &TransferMoneyCommandHandler{
		accounts:   accounts,
		users:      users,
		transfers:  transfers,
		authorizer: authorizer,
		publisher:  publisher,
	}
}

func (h *TransferMoneyCommandHandler) Handle(ctx context.Context, cmd TransferMoneyCommand) (*TransferMoneyResult, error) {
	payerID, err := uuid.Parse(cmd.PayerAccountID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPayerAccountID, err)
	}
	payeeID, err := uuid.Parse(cmd.PayeeAccountID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPayeeAccountID, err)
	}

	from, err := h.loadAccount(ctx, payerID, ErrPayerAccountNotFound)
	if err != nil {
		return nil, err
	}
	to, err := h.loadAccount(ctx, payeeID, ErrPayeeAccountNotFound)
	if err != nil {
		return nil, err
	}

	payer, err := h.users.ByID(ctx, from.OwnerId)
	if err != nil {
		return nil, fmt.Errorf("failed to load payer: %w", err)
	}

	transferID := uuid.New()
	amount := domain.MonetaryAmount{Currency: from.Balance.Currency, Value: cmd.Value}

	if err := h.authorizer.Authorize(ctx, transferID.String()); err != nil {
		transfer := account.NewFailedTransfer(transferID, from.ID, to.ID, amount, err.Error())
		if saveErr := h.transfers.Save(ctx, transfer, from, to); saveErr != nil {
			return nil, fmt.Errorf("failed to persist transfer: %w", saveErr)
		}
		if publishErr := h.publish(ctx, transfer); publishErr != nil {
			return nil, publishErr
		}
		return nil, fmt.Errorf("%w: %v", ErrTransferNotAuthorized, err)
	}

	transfer, domainErr := account.NewTransfer(transferID, from, to, payer, amount)
	if err := h.transfers.Save(ctx, transfer, from, to); err != nil {
		return nil, fmt.Errorf("failed to persist transfer: %w", err)
	}
	if err := h.publish(ctx, transfer); err != nil {
		return nil, err
	}
	if domainErr != nil {
		return nil, mapTransferError(domainErr)
	}

	return &TransferMoneyResult{
		TransferID:     transferID.String(),
		Status:         string(transfer.Status),
		Value:          transfer.Amount.Value,
		Currency:       transfer.Amount.Currency,
		PayerAccountID: uuid.UUID(from.ID).String(),
		PayeeAccountID: uuid.UUID(to.ID).String(),
		Events:         transfer.UncommittedEvents(),
	}, nil
}

func (h *TransferMoneyCommandHandler) loadAccount(ctx context.Context, id uuid.UUID, notFound error) (*account.Account, error) {
	acc, err := h.accounts.ByID(ctx, account.AccountID(id))
	if err != nil {
		var accountNotFound *account.AccountNotFoundError
		if errors.As(err, &accountNotFound) {
			return nil, fmt.Errorf("%w: %v", notFound, accountNotFound)
		}
		return nil, fmt.Errorf("failed to load account: %w", err)
	}
	return acc, nil
}

func (h *TransferMoneyCommandHandler) publish(ctx context.Context, transfer *account.Transfer) error {
	for _, event := range transfer.UncommittedEvents() {
		if err := h.publisher.Publish(ctx, event); err != nil {
			return fmt.Errorf("%w: %v", ErrTransferEventPublishFailed, err)
		}
	}
	return nil
}

func mapTransferError(err error) error {
	var sameAccount *account.SameAccountTransferError
	var currencyMismatch *account.CurrencyMismatchError
	var insufficientBalance *account.InsufficientBalanceError
	var notTransactable *account.AccountNotTransactableError
	var notAllowed *account.TransferNotAllowedError

	switch {
	case errors.As(err, &sameAccount):
		return fmt.Errorf("%w: %v", ErrSameAccountTransfer, sameAccount)
	case errors.As(err, &currencyMismatch):
		return fmt.Errorf("%w: %v", ErrCurrencyMismatch, currencyMismatch)
	case errors.As(err, &insufficientBalance):
		return fmt.Errorf("%w: %v", ErrInsufficientBalance, insufficientBalance)
	case errors.As(err, &notTransactable):
		return fmt.Errorf("%w: %v", ErrAccountNotTransactable, notTransactable)
	case errors.As(err, &notAllowed):
		return fmt.Errorf("%w: %v", ErrTransferNotAllowed, notAllowed)
	default:
		return err
	}
}
```

## T017 — Translate a missing account on the real read path

File: `internal/adapters/postgres/account_repository.go`
Test: `internal/adapters/postgres/account_repository_test.go` — `TestAccountRepository_ByIDNotFound` (T018)
Done: `go test ./internal/adapters/postgres/...` green; `ByID` returns `AccountNotFoundError` when the query finds no row and leaves every other error untouched

OLD:
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

NEW:
```go
func (r *AccountRepository) ByID(ctx context.Context, id account.AccountID) (*account.Account, error) {
	row := r.db.QueryRowContext(ctx, selectAccountByIDSQL, uuid.UUID(id))
	acc, err := scanAccount(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &account.AccountNotFoundError{AccountID: uuid.UUID(id).String()}
		}
		return nil, err
	}
	return acc, nil
}
```

## T018 — Test the missing account read path

File: `internal/adapters/postgres/account_repository_test.go`
Test: `go test ./internal/adapters/postgres/ -run TestAccountRepository_ByIDNotFound -v`
Done: the test passes and fails if the `sql.ErrNoRows` translation is removed

OLD:
```go
func TestAccountRepository_AllAccounts(t *testing.T) {
```

NEW:
```go
func TestAccountRepository_ByIDNotFound(t *testing.T) {
	db, mock := newMockDB(t)
	repo := NewAccountRepository(db)
	_, id, _ := testAccount(t)

	mock.ExpectQuery(`SELECT id, owner_id, currency, balance, status, created_at`).
		WithArgs(id.String()).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.ByID(context.Background(), account.AccountID(id))
	var notFound *account.AccountNotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("expected AccountNotFoundError, got %v", err)
	}
}

func TestAccountRepository_AllAccounts(t *testing.T) {
```

## T019 — Make the existing account fake signal not-found

File: `internal/application/account/command/create_account_command_handler_test.go`
Test: `internal/application/account/command/` full package run after T020
Done: the fake returns `AccountNotFoundError` for an unknown id; `go test ./internal/application/account/command/...` green; the create-account tests are unchanged and still pass

OLD:
```go
func (f *fakeAccountRepository) ByID(_ context.Context, id account.AccountID) (*account.Account, error) {
	return f.accounts[uuid.UUID(id)], nil
}
```

NEW:
```go
func (f *fakeAccountRepository) ByID(_ context.Context, id account.AccountID) (*account.Account, error) {
	acc, ok := f.accounts[uuid.UUID(id)]
	if !ok {
		return nil, &account.AccountNotFoundError{AccountID: uuid.UUID(id).String()}
	}
	return acc, nil
}
```

## T020 — Test the transfer use case with fakes

File: `internal/application/account/command/transfer_money_command_handler_test.go`
Test: `go test ./internal/application/account/command/ -run TestTransferMoneyCommandHandler -v`
Done: all nine tests pass; the not-authorized and domain-refusal tests assert unchanged balances, exactly one saved FAILED transfer and the published `AccountTransferCreated` + `AccountTransferFailed` pair; the publisher-failure test asserts the attempt is never reported as success

OLD:
```
(file does not exist)
```

NEW:
```go
package command

import (
	"context"
	"errors"
	"testing"

	"event-driven-architecture/internal/domain"
	"event-driven-architecture/internal/domain/account"
	"event-driven-architecture/internal/domain/user"

	"github.com/google/uuid"
)

type stubHasher struct{}

func (stubHasher) Hash(plain string) (string, error) {
	return "hashed:" + plain, nil
}

type fakeUserRepository struct {
	users map[uuid.UUID]*user.User
}

func (f *fakeUserRepository) Add(_ context.Context, u *user.User) error {
	f.users[uuid.UUID(u.ID)] = u
	return nil
}

func (f *fakeUserRepository) ByID(_ context.Context, id user.ID) (*user.User, error) {
	u, ok := f.users[uuid.UUID(id)]
	if !ok {
		return nil, errors.New("user not found")
	}
	return u, nil
}

func (f *fakeUserRepository) AllUsers(_ context.Context) ([]*user.User, error) {
	out := make([]*user.User, 0, len(f.users))
	for _, u := range f.users {
		out = append(out, u)
	}
	return out, nil
}

type fakeTransferRepository struct {
	saved         []*account.Transfer
	payerBalances []float64
	payeeBalances []float64
	saveErr       error
}

func (f *fakeTransferRepository) ByID(_ context.Context, id string) (*account.Transfer, error) {
	return nil, nil
}

func (f *fakeTransferRepository) Save(_ context.Context, transfer *account.Transfer, from, to *account.Account) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.saved = append(f.saved, transfer)
	f.payerBalances = append(f.payerBalances, from.Balance.Value)
	f.payeeBalances = append(f.payeeBalances, to.Balance.Value)
	return nil
}

func (f *fakeTransferRepository) AllTransfers(_ context.Context) ([]*account.Transfer, error) {
	return f.saved, nil
}

type stubAuthorizer struct {
	err error
}

func (s stubAuthorizer) Authorize(_ context.Context, _ string) error {
	return s.err
}

type recordingPublisher struct {
	events []domain.Event
	err    error
}

func (p *recordingPublisher) Publish(_ context.Context, event domain.Event) error {
	if p.err != nil {
		return p.err
	}
	p.events = append(p.events, event)
	return nil
}

func newTransferHandlerFixture(t *testing.T) (*fakeAccountRepository, *fakeUserRepository, *account.Account, *account.Account) {
	t.Helper()
	ownerID := user.ID(uuid.New())

	from, err := account.NewAccount(account.AccountID(uuid.New()), ownerID, domain.MonetaryAmount{Currency: "BRL", Value: 100})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	to, err := account.NewAccount(account.AccountID(uuid.New()), user.ID(uuid.New()), domain.MonetaryAmount{Currency: "BRL", Value: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	payer, err := user.NewUser(ownerID, stubHasher{}, "Rita Fontenele", "52998224725", "rita@example.com", "secret", user.COMMON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	accounts := newFakeAccountRepository()
	accounts.accounts[uuid.UUID(from.ID)] = from
	accounts.accounts[uuid.UUID(to.ID)] = to
	users := &fakeUserRepository{users: map[uuid.UUID]*user.User{uuid.UUID(payer.ID): payer}}

	return accounts, users, from, to
}

func assertPublishedEvents(t *testing.T, publisher *recordingPublisher, names ...string) {
	t.Helper()
	if len(publisher.events) != len(names) {
		t.Fatalf("expected %d published events, got %d", len(names), len(publisher.events))
	}
	for i, name := range names {
		if publisher.events[i].EventName() != name {
			t.Errorf("event %d: expected %s, got %s", i, name, publisher.events[i].EventName())
		}
	}
}

func TestTransferMoneyCommandHandler_Success(t *testing.T) {
	accounts, users, from, to := newTransferHandlerFixture(t)
	transfers := &fakeTransferRepository{}
	publisher := &recordingPublisher{}
	h := NewTransferMoneyCommandHandler(accounts, users, transfers, stubAuthorizer{}, publisher)

	result, err := h.Handle(context.Background(), TransferMoneyCommand{
		PayerAccountID: uuid.UUID(from.ID).String(),
		PayeeAccountID: uuid.UUID(to.ID).String(),
		Value:          40,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TransferID == "" {
		t.Error("expected a transfer id")
	}
	if result.Status != string(account.COMPLETED) {
		t.Errorf("expected COMPLETED, got %s", result.Status)
	}
	if result.Value != 40 || result.Currency != "BRL" {
		t.Errorf("unexpected result: %+v", result)
	}
	if from.Balance.Value != 60 || to.Balance.Value != 40 {
		t.Errorf("expected balances 60 and 40, got %v and %v", from.Balance.Value, to.Balance.Value)
	}
	if len(transfers.saved) != 1 {
		t.Fatalf("expected 1 saved transfer, got %d", len(transfers.saved))
	}
	if len(result.Events) != 2 {
		t.Errorf("expected 2 events in the result, got %d", len(result.Events))
	}
	assertPublishedEvents(t, publisher, "AccountTransferCreated", "AccountTransferSucceeded")
}

func TestTransferMoneyCommandHandler_InvalidPayerAccountID(t *testing.T) {
	accounts, users, _, _ := newTransferHandlerFixture(t)
	transfers := &fakeTransferRepository{}
	publisher := &recordingPublisher{}
	h := NewTransferMoneyCommandHandler(accounts, users, transfers, stubAuthorizer{}, publisher)

	_, err := h.Handle(context.Background(), TransferMoneyCommand{
		PayerAccountID: "not-a-uuid",
		PayeeAccountID: uuid.New().String(),
		Value:          10,
	})
	if !errors.Is(err, ErrInvalidPayerAccountID) {
		t.Fatalf("expected ErrInvalidPayerAccountID, got %v", err)
	}
	if len(transfers.saved) != 0 || len(publisher.events) != 0 {
		t.Error("expected no port to be called")
	}
}

func TestTransferMoneyCommandHandler_InvalidPayeeAccountID(t *testing.T) {
	accounts, users, _, _ := newTransferHandlerFixture(t)
	transfers := &fakeTransferRepository{}
	publisher := &recordingPublisher{}
	h := NewTransferMoneyCommandHandler(accounts, users, transfers, stubAuthorizer{}, publisher)

	_, err := h.Handle(context.Background(), TransferMoneyCommand{
		PayerAccountID: uuid.New().String(),
		PayeeAccountID: "not-a-uuid",
		Value:          10,
	})
	if !errors.Is(err, ErrInvalidPayeeAccountID) {
		t.Fatalf("expected ErrInvalidPayeeAccountID, got %v", err)
	}
	if len(transfers.saved) != 0 || len(publisher.events) != 0 {
		t.Error("expected no port to be called")
	}
}

func TestTransferMoneyCommandHandler_PayerAccountNotFound(t *testing.T) {
	accounts, users, _, to := newTransferHandlerFixture(t)
	transfers := &fakeTransferRepository{}
	publisher := &recordingPublisher{}
	h := NewTransferMoneyCommandHandler(accounts, users, transfers, stubAuthorizer{}, publisher)

	_, err := h.Handle(context.Background(), TransferMoneyCommand{
		PayerAccountID: uuid.New().String(),
		PayeeAccountID: uuid.UUID(to.ID).String(),
		Value:          10,
	})
	if !errors.Is(err, ErrPayerAccountNotFound) {
		t.Fatalf("expected ErrPayerAccountNotFound, got %v", err)
	}
	if len(transfers.saved) != 0 || len(publisher.events) != 0 {
		t.Error("expected nothing to be saved or published")
	}
}

func TestTransferMoneyCommandHandler_PayeeAccountNotFound(t *testing.T) {
	accounts, users, from, _ := newTransferHandlerFixture(t)
	transfers := &fakeTransferRepository{}
	publisher := &recordingPublisher{}
	h := NewTransferMoneyCommandHandler(accounts, users, transfers, stubAuthorizer{}, publisher)

	_, err := h.Handle(context.Background(), TransferMoneyCommand{
		PayerAccountID: uuid.UUID(from.ID).String(),
		PayeeAccountID: uuid.New().String(),
		Value:          10,
	})
	if !errors.Is(err, ErrPayeeAccountNotFound) {
		t.Fatalf("expected ErrPayeeAccountNotFound, got %v", err)
	}
	if len(transfers.saved) != 0 || len(publisher.events) != 0 {
		t.Error("expected nothing to be saved or published")
	}
}

func TestTransferMoneyCommandHandler_NotAuthorized(t *testing.T) {
	accounts, users, from, to := newTransferHandlerFixture(t)
	transfers := &fakeTransferRepository{}
	publisher := &recordingPublisher{}
	h := NewTransferMoneyCommandHandler(accounts, users, transfers, stubAuthorizer{err: errors.New("gateway denied")}, publisher)

	_, err := h.Handle(context.Background(), TransferMoneyCommand{
		PayerAccountID: uuid.UUID(from.ID).String(),
		PayeeAccountID: uuid.UUID(to.ID).String(),
		Value:          40,
	})
	if !errors.Is(err, ErrTransferNotAuthorized) {
		t.Fatalf("expected ErrTransferNotAuthorized, got %v", err)
	}
	if from.Balance.Value != 100 || to.Balance.Value != 0 {
		t.Errorf("expected the balances to be unchanged, got %v and %v", from.Balance.Value, to.Balance.Value)
	}
	if len(transfers.saved) != 1 {
		t.Fatalf("expected the attempt to be recorded, got %d saved transfers", len(transfers.saved))
	}
	if transfers.saved[0].Status != account.FAILED {
		t.Errorf("expected a FAILED transfer, got %s", transfers.saved[0].Status)
	}
	assertPublishedEvents(t, publisher, "AccountTransferCreated", "AccountTransferFailed")
}

func TestTransferMoneyCommandHandler_DomainRefusal(t *testing.T) {
	accounts, users, from, to := newTransferHandlerFixture(t)
	transfers := &fakeTransferRepository{}
	publisher := &recordingPublisher{}
	h := NewTransferMoneyCommandHandler(accounts, users, transfers, stubAuthorizer{}, publisher)

	_, err := h.Handle(context.Background(), TransferMoneyCommand{
		PayerAccountID: uuid.UUID(from.ID).String(),
		PayeeAccountID: uuid.UUID(to.ID).String(),
		Value:          500,
	})
	if !errors.Is(err, ErrInsufficientBalance) {
		t.Fatalf("expected ErrInsufficientBalance, got %v", err)
	}
	if from.Balance.Value != 100 || to.Balance.Value != 0 {
		t.Errorf("expected the balances to be unchanged, got %v and %v", from.Balance.Value, to.Balance.Value)
	}
	if len(transfers.saved) != 1 || transfers.saved[0].Status != account.FAILED {
		t.Error("expected exactly one saved FAILED transfer")
	}
	assertPublishedEvents(t, publisher, "AccountTransferCreated", "AccountTransferFailed")
}

func TestTransferMoneyCommandHandler_SaveFailure(t *testing.T) {
	accounts, users, from, to := newTransferHandlerFixture(t)
	transfers := &fakeTransferRepository{saveErr: errors.New("db down")}
	publisher := &recordingPublisher{}
	h := NewTransferMoneyCommandHandler(accounts, users, transfers, stubAuthorizer{}, publisher)

	_, err := h.Handle(context.Background(), TransferMoneyCommand{
		PayerAccountID: uuid.UUID(from.ID).String(),
		PayeeAccountID: uuid.UUID(to.ID).String(),
		Value:          40,
	})
	if err == nil {
		t.Fatal("expected an error")
	}
	if errors.Is(err, ErrInsufficientBalance) {
		t.Error("a persistence failure must not be reported as a domain refusal")
	}
	if len(publisher.events) != 0 {
		t.Error("expected no event to be published when persisting fails")
	}
}

func TestTransferMoneyCommandHandler_PublisherFailure(t *testing.T) {
	accounts, users, from, to := newTransferHandlerFixture(t)
	transfers := &fakeTransferRepository{}
	publisher := &recordingPublisher{err: errors.New("broker down")}
	h := NewTransferMoneyCommandHandler(accounts, users, transfers, stubAuthorizer{}, publisher)

	result, err := h.Handle(context.Background(), TransferMoneyCommand{
		PayerAccountID: uuid.UUID(from.ID).String(),
		PayeeAccountID: uuid.UUID(to.ID).String(),
		Value:          40,
	})
	if !errors.Is(err, ErrTransferEventPublishFailed) {
		t.Fatalf("expected ErrTransferEventPublishFailed, got %v", err)
	}
	if result != nil {
		t.Error("expected no result when publishing fails")
	}
}

var _ TransferMoneyUseCase = (*TransferMoneyCommandHandler)(nil)

var _ account.TransferRepository = (*fakeTransferRepository)(nil)

var _ account.TransferAuthorizer = stubAuthorizer{}

var _ user.UserRepository = (*fakeUserRepository)(nil)

var _ TransferEventPublisher = (*recordingPublisher)(nil)
```

## T021 — Record the decision

File: `.ai/decisions/0008-logical-transfer-slice.md`
Test: not applicable (documentation) — verified by `go test ./...` staying green
Done: the ADR exists with `Status: accepted`, lists the decision and the consequences, and its "Known gaps" list matches the plan's non-goals

OLD:
```
(file does not exist)
```

NEW:
```markdown
# ADR-0008: Logical transfer slice — domain factory owns the rules, events are collected in the domain

Status: accepted (2026-09-20)

## Context
`POST /transfer` is a documented but unbuilt requirement. The domain had a transfer flow
(`NewTransfer` struct holding repositories plus a `TransferService`) that never loaded the
payee, never validated the amount, returned untyped errors and never raised the
`AccountTransfer*` events it declared. Persistence, authorization and notification
infrastructure do not exist in the repository.

## Decision
- The slice is **logical only**: domain rules + domain events + one application use case. No
  HTTP route, no Kafka, no event store, no balance persistence, no notification.
- **Domain factory** `account.NewTransfer(id, from, to, payer, amount)` enforces every rule
  (amount positive, payer different from payee, same currency, payer is COMMON, account ACTIVE,
  sufficient balance) and moves the money through `Account.Debit`/`Account.Credit`.
  `HasBalance` compares with `>=`, so transferring the whole balance is legal.
- A refused attempt returns a **non-nil FAILED transfer together with the typed error**: the
  refused attempt is the audit record (`AccountTransferCreated` + `AccountTransferFailed`).
- **Uncommitted events**: the domain collects events on the transfer (`UncommittedEvents()`) and
  the application handler publishes them through the `TransferEventPublisher` port. The domain
  never holds a publisher.
- `domain.Event` (`EventName() string`) is introduced in the root package and implemented by the
  three transfer events only; the remaining events are a follow-up.
- **Authorization** is an outbound port (`account.TransferAuthorizer`) called by the handler
  before any balance change; `NewFailedTransfer` records the refusal. The real gateway adapter
  is deferred.
- **Single write seam**: `TransferRepository.Add` was replaced by
  `Save(ctx, transfer, from, to)` — one call per attempt, and the place where the future
  Postgres adapter must be atomic (DDD-07). No balance is persisted in this slice.
- `TransferService` and the `NewTransfer` struct were **deleted** (dead after the rework).
- Business-rule refusals are mapped to application sentinels (`ErrInsufficientBalance`,
  `ErrSameAccountTransfer`, `ErrCurrencyMismatch`, `ErrAccountNotTransactable`,
  `ErrTransferNotAllowed`); field violations keep flowing as `ConstraintValidationError`.
- `AccountRepository.ByID` translates `sql.ErrNoRows` into `AccountNotFoundError`, so the
  not-found contract holds for the real adapter and not only for the fakes.

## Consequences
- The transfer rules live in one place and are proven by pure domain tests; the handler has no
  rule branch.
- The service still does not expose transfers: nothing is wired in `main.go` and only the read
  path has a real implementation.
- Known gaps: balance persistence and its atomicity ADR, concurrency safety, idempotency,
  authorization adapter, notification + DLQ, broker publishing/event store, `EventName()` for
  the remaining events, migrating `NewDeposit`/`DepositService` to this pattern.
```

## T022 — Update the domain facts in memory

File: `.ai/architecture/domain.md`
Test: not applicable (documentation) — verified by `go test ./...` staying green
Done: the ports paragraph and the business-rules paragraph describe `Save`, the new ports, `Debit`/`Credit`, the factories, the collected events and the typed errors; no mention of `TransferService` or `NewTransfer.CreateTransfer` remains in the file

OLD:
```
- Ports: `AccountRepository` (ByID/Add/AllAccounts), `DepositRepository`, `TransferRepository` (note: Deposit/Transfer repos use `id string`, not a typed ID), `AccountQueries` — read-side port `OwnerBalance(ctx, ownerID uuid.UUID)`.
- Domain services: `DepositService`, `TransferService` interfaces (adapter impls not yet written).
- Unusual pattern: `NewDeposit` / `NewTransfer` are structs that hold injected deps (service + repositories) and expose `Create()` / `CreateTransfer(ctx, ...)`. Callers must wire the deps; no plain constructor exists for these.
```

NEW:
```
- Ports: `AccountRepository` (ByID/Add/AllAccounts), `DepositRepository`, `TransferRepository` (`ByID`/`Save(ctx, transfer, from, to)`/`AllTransfers` — `Save` is the single write seam, the future transaction boundary; Deposit repo still uses `id string`), `AccountQueries` — read-side port `OwnerBalance(ctx, ownerID uuid.UUID)`, `TransferAuthorizer` (`Authorize(ctx, transferID) error`) — outbound authorization port next to the rule it serves (ADR-0008).
- Domain services: `DepositService` interface (no adapter implementation). `TransferService` was **deleted** in ADR-0008 — the transfer flow is now a factory.
- Factories (ADR-0008): `NewTransfer(id, from, to, payer, amount) (*Transfer, error)` validates and executes a transfer, returning a non-nil FAILED transfer together with the typed error on refusal; `NewFailedTransfer(id, from, to, amount, reason)` records an attempt refused before execution (authorization). Both collect `AccountTransferCreated` and then `AccountTransferSucceeded`/`AccountTransferFailed`; `Transfer.UncommittedEvents()` exposes them.
- `Account.Debit(amount) error` (refuses a non-ACTIVE account and an insufficient balance) and `Account.Credit(amount)`; `HasBalance(amount)` compares with `>=`.
- `domain.Event` (`EventName() string`) lives in the root package; implemented by the three transfer events only.
- Unusual pattern that remains: `NewDeposit` is a struct holding injected deps (`DepositService`) and exposing `Create()`; migrating it is a follow-up.
```

OLD:
```
## Business rules encoded in domain
- Only COMMON users may initiate transfers (`User.CanTransfer`); SHOPKEEPER receives only.
- Sender account must be ACTIVE and have sufficient balance (`NewTransfer.CreateTransfer`).
```

NEW:
```
## Business rules encoded in domain
- Only COMMON users may initiate transfers (`User.CanTransfer`); SHOPKEEPER receives only.
- Sender account must be ACTIVE and have a sufficient balance (`Account.Debit`, called by `NewTransfer`); an amount equal to the balance is sufficient.
- A transfer amount must be positive (`domain.ConstraintValidationError{Field: "amount"}`), the payer and payee accounts must differ and must hold the same currency, and no conversion is performed.
- Every attempt produces `created` then exactly one outcome event; refusals carry the reason in `AccountTransferFailed`.
```

## T023 — Update the application facts in memory

File: `.ai/architecture/application.md`
Test: not applicable (documentation) — verified by `go test ./...` staying green
Done: the account commands section documents the transfer use case, its five ports, the authorization ordering, the single `Save` call and the sentinel mapping; the sentinel list is complete

OLD:
```
- Sentinel errors in `errors.go`: `ErrAccountAlreadyExists`, `ErrOwnerNotFound`,
  `ErrInvalidOwnerID` (plain `errors.New`, compared with `errors.Is`).
```

NEW:
```
- Sentinel errors in `errors.go`: `ErrAccountAlreadyExists`, `ErrOwnerNotFound`,
  `ErrInvalidOwnerID` (plain `errors.New`, compared with `errors.Is`), plus the transfer
  sentinels `ErrInvalidPayerAccountID`, `ErrInvalidPayeeAccountID`, `ErrPayerAccountNotFound`,
  `ErrPayeeAccountNotFound`, `ErrTransferNotAuthorized`, `ErrInsufficientBalance`,
  `ErrAccountNotTransactable`, `ErrTransferNotAllowed`, `ErrSameAccountTransfer`,
  `ErrCurrencyMismatch`, `ErrTransferEventPublishFailed`.
- `TransferMoneyCommand` `{PayerAccountID, PayeeAccountID string, Value float64}` +
  `TransferMoneyCommandHandler` (ADR-0008) implements the `TransferMoneyUseCase` port
  (`Handle(ctx, cmd) (*TransferMoneyResult, error)`), the port the future HTTP adapter consumes.
  Constructor ports: `account.AccountRepository`, `user.UserRepository`,
  `account.TransferRepository`, `account.TransferAuthorizer`, `TransferEventPublisher`.
  Order: parse both UUIDs → load payer account, payee account, payer user → authorize (a refusal
  records a FAILED transfer through `account.NewFailedTransfer`, persists, publishes and returns
  `ErrTransferNotAuthorized` with no balance change) → `account.NewTransfer` (the domain enforces
  every rule and collects the events) → one `transfers.Save(ctx, transfer, from, to)` call per
  attempt → publish each `UncommittedEvents()` entry (failure ⇒
  `ErrTransferEventPublishFailed`, never success) → map domain refusals to sentinels, leaving
  `domain.ConstraintValidationError` untouched (→ 400).
- `TransferMoneyResult` `{TransferID, Status, Value, Currency, PayerAccountID, PayeeAccountID,
  Events []domain.Event}`. `TransferEventPublisher` is the outbound publisher port (only the test
  fake implements it).
- The use case is not wired into `main.go`: no adapter implements its write, authorization or
  publisher ports yet.
```

## T024 — Update the adapter facts in memory

File: `.ai/architecture/adapters.md`
Test: not applicable (documentation) — verified by `go test ./...` staying green
Done: the Postgres paragraph mentions the `sql.ErrNoRows` → `AccountNotFoundError` translation

OLD:
```
- `account_repository.go` — implements the `account.AccountRepository` port. SQL statements are package-level consts; `scanAccount` scans via a small `rowScanner` interface (works for both `*sql.Row` and `*sql.Rows`); `translateError` maps `pgconn.PgError` codes → domain error types: `23505` unique_violation → `AccountAlreadyExistsError`, `23503` foreign_key_violation → `OwnerNotFoundError`, anything else passes through.
```

NEW:
```
- `account_repository.go` — implements the `account.AccountRepository` port. SQL statements are package-level consts; `scanAccount` scans via a small `rowScanner` interface (works for both `*sql.Row` and `*sql.Rows`); `translateError` maps `pgconn.PgError` codes → domain error types: `23505` unique_violation → `AccountAlreadyExistsError`, `23503` foreign_key_violation → `OwnerNotFoundError`, anything else passes through; `ByID` translates `sql.ErrNoRows` → `AccountNotFoundError` (ADR-0008).
```

## T025 — Update the development context

File: `.ai/context/development.md`
Test: not applicable (documentation) — verified by `go test ./...` staying green
Done: the state/gaps section describes the logical transfer slice and no longer calls the transfer handlers empty stubs

OLD:
```
## State / known gaps
- Event sourcing defined, but no persistence/replay.
- Transfer/deposit command handlers are empty stubs (`internal/application/user/command`).
```

NEW:
```
## State / known gaps
- Event sourcing defined, but no persistence/replay.
- The transfer flow is implemented as a logical slice (ADR-0008): domain rules, collected
  domain events and `TransferMoneyCommandHandler` over port fakes. Not wired into the running
  service; no balance persistence, no authorization adapter, no notification, no HTTP route.
- `internal/application/user/command/user_deposit_money_*` and `user_transfer_money_*` remain
  empty stubs (superseded by the account transfer use case); deposit has no use case yet.
- `internal/domain/account/transfer_service.go` was deleted (ADR-0008).
```

## T026 — Update the testing context

File: `.ai/context/testing.md`
Test: not applicable (documentation) — the recorded status must match the output of `go test ./...`
Done: the status section records the transfer slice tests, and the date matches the day the verification command was actually run

OLD:
```
## Status (verified 2026-08-29)
- `go test ./...` passes. Coverage: domain/account, domain/user, application/account/command, application/user/command, adapters/http, adapters/postgres, adapters/crypto, config.
- No tests yet for: user queries/projections, transfer/deposit flows.
```

NEW:
```
## Status (verified 2026-09-20)
- `go test ./...` passes. Coverage: domain/account (including the transfer rules and the
  collected domain events), domain/user, application/account/command (create account, transfer
  with hand-rolled port fakes), application/user/command, adapters/http, adapters/postgres,
  adapters/crypto, config.
- Transfer flow tests: `internal/domain/account/transfer_test.go` (rules + event order) and
  `internal/application/account/command/transfer_money_command_handler_test.go` (fakes:
  `fakeTransferRepository`, `fakeUserRepository`, `stubAuthorizer`, `recordingPublisher`).
- No tests yet for: user queries/projections, the deposit flow, the HTTP transfer route.
```
