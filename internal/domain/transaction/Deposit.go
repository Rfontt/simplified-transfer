package transaction

import "event-driven-architecture/internal/domain/user"

type Deposit struct {
	Transaction Transaction
	UserID      user.ID
}

type NewDeposit struct {
	userId             user.ID
	transaction        Transaction
	service            DepositService
	transactionService TransactionService
}

func (deposit NewDeposit) Create() error {
	err := deposit.service.Create(
		Deposit{
			Transaction: deposit.transaction,
			UserID:      deposit.userId,
		},
	)

	// TODO(rfontt): if it throw an error so emit deposit event error

	err = deposit.transactionService.Create(deposit.transaction)

	// TODO(rfontt): if it throw an error so emit transaction event error

	if err != nil {
		return err
	}

	return nil
}
