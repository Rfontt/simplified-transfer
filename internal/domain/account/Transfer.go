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
	userService       user.UserService
	transferService   TransferService
	accountRepository AccountRepository
}

func (t NewTransfer) CreateTransfer(ctx context.Context, from, to AccountID, amount domain.MonetaryAmount) error {
	fromAccount, err := t.accountRepository.ByID(ctx, from)
	if err != nil {
		return err
	}

	if !fromAccount.CanTransact() || !fromAccount.HasBalance(amount) {
		return errors.New("sender account is not allowed to perform transfers")
	}

	userFrom, err := t.userService.GetOne(fromAccount.OwnerId)
	if err != nil {
		return err
	}

	if userFrom.Type == user.SHOPKEEPER {
		return fmt.Errorf("%v user type is not allowed", userFrom.Type)
	}

	if err := t.transferService.Create(from, to, amount); err != nil {
		return err
	}

	// TODO: emit domain events here

	return nil
}
