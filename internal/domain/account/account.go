package account

import (
	"strings"

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

func NewAccount(id AccountID, ownerID user.ID, balance domain.MonetaryAmount) (*Account, error) {
	currency := strings.TrimSpace(balance.Currency)
	if currency == "" {
		return nil, &InvalidCurrencyError{}
	}
	if balance.Value < 0 {
		return nil, &InvalidBalanceError{}
	}
	return &Account{
		ID:        id,
		OwnerId:   ownerID,
		Balance:   domain.MonetaryAmount{Currency: currency, Value: balance.Value},
		Status:    ACTIVE,
		CreatedAt: time.Now(),
	}, nil
}

func (a Account) CanTransact() bool {
	return a.Status == ACTIVE
}

func (a Account) HasBalance(amount domain.MonetaryAmount) bool {
	return a.Balance.Value > amount.Value
}
