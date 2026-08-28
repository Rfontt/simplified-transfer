package account

import (
	"testing"

	"event-driven-architecture/internal/domain"
	"event-driven-architecture/internal/domain/user"

	"github.com/google/uuid"
)

func TestNewAccount(t *testing.T) {
	id := AccountID(uuid.New())
	ownerID := user.ID(uuid.New())
	balance := domain.MonetaryAmount{Currency: "BRL", Value: 100}

	acc := NewAccount(id, ownerID, balance)

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
