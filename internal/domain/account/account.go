package account

import (
	"event-driven-architecture/internal/domain"
	"event-driven-architecture/internal/domain/user"
	"time"
)

type AccountStatus string

const (
	ACTIVE  AccountStatus = "active"
	BLOCKED AccountStatus = "blocked"
	CLOSED  AccountStatus = "closed"
)

type AccountID domain.AggregateID

type Account struct {
	ID        AccountID
	OwnerId   user.ID
	Balance   domain.MonetaryAmount
	Status    AccountStatus
	CreatedAt time.Time
}

func NewAccount(id AccountID, ownerID user.ID, balance domain.MonetaryAmount) *Account {
	return &Account{
		ID:        id,
		OwnerId:   ownerID,
		Balance:   balance,
		Status:    ACTIVE,
		CreatedAt: time.Now(),
	}
}

func (a Account) CanTransact() bool {
	return a.Status == ACTIVE
}

func (a Account) HasBalance(amount domain.MonetaryAmount) bool {
	return a.Balance.Value > amount.Value
}
