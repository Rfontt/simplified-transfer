package account

import (
	"errors"
	"testing"

	"event-driven-architecture/internal/domain"
	"event-driven-architecture/internal/domain/user"

	"github.com/google/uuid"
)

func TestNewAccount(t *testing.T) {
	id := AccountID(uuid.New())
	ownerID := user.ID(uuid.New())
	balance := domain.MonetaryAmount{Currency: "BRL", Value: 100}

	acc, err := NewAccount(id, ownerID, balance)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if acc.ID != id {
		t.Errorf("expected ID %v, got %v", id, acc.ID)
	}
	if acc.OwnerId != ownerID {
		t.Errorf("expected OwnerId %v, got %v", ownerID, acc.OwnerId)
	}
	if acc.Balance != balance {
		t.Errorf("expected Balance %v, got %v", balance, acc.Balance)
	}
	if acc.Status != ACTIVE {
		t.Errorf("expected status ACTIVE, got %v", acc.Status)
	}
	if acc.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestNewAccount_InvalidBalance(t *testing.T) {
	id := AccountID(uuid.New())
	ownerID := user.ID(uuid.New())

	for _, currency := range []string{"", "  "} {
		_, err := NewAccount(id, ownerID, domain.MonetaryAmount{Currency: currency, Value: 100})
		assertConstraintError(t, err, "currency")
	}

	_, err := NewAccount(id, ownerID, domain.MonetaryAmount{Currency: "BRL", Value: -1})
	assertConstraintError(t, err, "balance")
}

func assertConstraintError(t *testing.T, err error, field string) {
	t.Helper()
	var invalid *domain.ConstraintValidationError
	if !errors.As(err, &invalid) {
		t.Fatalf("expected ConstraintValidationError, got %v", err)
	}
	if invalid.Field != field {
		t.Errorf("expected field %q, got %q", field, invalid.Field)
	}
}

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
