package account

import (
	"errors"
	"event-driven-architecture/internal/domain"

	"github.com/google/uuid"
)

type Deposit struct {
	ID        uuid.UUID
	AccountId AccountID
	Amount    domain.MonetaryAmount
	Status    AccountTransactionStatus
}

type NewDeposit struct {
	accountId AccountID
	amount    domain.MonetaryAmount
	service   DepositService
}

func (deposit NewDeposit) Create() error {

	if deposit.amount.Value <= 0 {
		return errors.New("amount must be greater than zero")
	}

	err := deposit.service.Create(
		deposit.accountId,
		deposit.amount,
	)

	// TODO(rfontt): if it throw an error so emit deposit event error

	if err != nil {
		return err
	}

	return nil
}
