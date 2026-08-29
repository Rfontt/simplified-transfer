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
