package account

import (
	"errors"
	"event-driven-architecture/internal/domain"
	"event-driven-architecture/internal/domain/user"
	"fmt"

	"github.com/google/uuid"
)

type TransferStatus string

const (
	PENDING   TransferStatus = "PENDING"
	COMPLETED TransferStatus = "COMPLETED"
	FAILED    TransferStatus = "FAILED"
)

type Transfer struct {
	ID     uuid.UUID
	From   AccountID
	To     AccountID
	Amount domain.MonetaryAmount
	Status TransferStatus
}

type NewTransfer struct {
	userService     user.UserService
	transferService TransferService
	accountService  AccountService
}

func (t NewTransfer) CreateTransfer(from, to AccountID, amount domain.MonetaryAmount) error {
	fromAccount, err := t.accountService.GetOne(from)
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
		return fmt.Errorf("%t user type is not allowed", userFrom.Type)
	}

	if err := t.transferService.Create(from, to, amount); err != nil {
		return err
	}

	// TODO: emit domain events here

	return nil
}
