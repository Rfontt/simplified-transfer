package account

import (
	"context"
	"errors"
	"event-driven-architecture/internal/domain"
	"event-driven-architecture/internal/domain/user"
	"fmt"

	"github.com/google/uuid"
)

type Transfer struct {
	ID     uuid.UUID
	From   AccountID
	To     AccountID
	Amount domain.MonetaryAmount
	Status AccountTransactionStatus
}

type NewTransfer struct {
	userRepository    user.UserRepository
	transferService   TransferService
	accountRepository AccountRepository
}

func (t NewTransfer) CreateTransfer(ctx context.Context, from, to AccountID, amount domain.MonetaryAmount) (*Transfer, error) {
	fromAccount, err := t.accountRepository.ByID(ctx, from)
	if err != nil {
		return nil, err
	}

	if !fromAccount.CanTransact() || !fromAccount.HasBalance(amount) {
		return nil, errors.New("sender account is not allowed to perform transfers")
	}

	userFrom, err := t.userRepository.ByID(ctx, fromAccount.OwnerId)
	if err != nil {
		return nil, err
	}

	if !userFrom.CanTransfer() {
		return nil, fmt.Errorf("%v user type is not allowed to transfer", userFrom.Type)
	}

	result, err := t.transferService.TransferFunds(from, to, amount)
	if err != nil {
		return nil, err
	}

	// TODO: emit domain events here

	return result, nil
}
