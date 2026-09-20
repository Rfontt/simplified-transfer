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
