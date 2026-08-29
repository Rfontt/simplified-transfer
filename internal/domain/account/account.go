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
	a := &Account{
		ID:        id,
		OwnerId:   ownerID,
		Balance:   balance,
		Status:    ACTIVE,
		CreatedAt: time.Now(),
	}
	if err := a.validateFields(); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *Account) validateFields() error {
	currency := strings.TrimSpace(a.Balance.Currency)
	if currency == "" {
		return &domain.ConstraintValidationError{Field: "currency"}
	}
	if a.Balance.Value < 0 {
		return &domain.ConstraintValidationError{Field: "balance"}
	}
	a.Balance = domain.MonetaryAmount{Currency: currency, Value: a.Balance.Value}
	return nil
}

func (a Account) CanTransact() bool {
	return a.Status == ACTIVE
}

func (a Account) HasBalance(amount domain.MonetaryAmount) bool {
	return a.Balance.Value > amount.Value
}
