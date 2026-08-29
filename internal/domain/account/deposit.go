package account

import (
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

func (deposit NewDeposit) Create() (*Deposit, error) {
	if !deposit.amount.IsPositive() {
		return nil, &domain.ConstraintValidationError{Field: "amount"}
	}

	result, err := deposit.service.DepositFunds(
		deposit.accountId,
		deposit.amount,
	)

	// TODO(rfontt): if it throw an error so emit deposit event error

	if err != nil {
		return nil, err
	}

	return result, nil
}
