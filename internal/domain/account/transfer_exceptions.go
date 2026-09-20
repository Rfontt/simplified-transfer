package account

import (
	"fmt"

	"github.com/google/uuid"
)

type SameAccountTransferError struct {
	AccountID AccountID
}

func (e *SameAccountTransferError) Error() string {
	return fmt.Sprintf("account %s cannot transfer to itself", uuid.UUID(e.AccountID))
}

type CurrencyMismatchError struct {
	FromCurrency string
	ToCurrency   string
}

func (e *CurrencyMismatchError) Error() string {
	return fmt.Sprintf("cannot transfer between accounts with different currencies: %s and %s", e.FromCurrency, e.ToCurrency)
}

type InsufficientBalanceError struct {
	AccountID AccountID
	Balance   float64
	Amount    float64
}

func (e *InsufficientBalanceError) Error() string {
	return fmt.Sprintf("account %s has insufficient balance: %v < %v", uuid.UUID(e.AccountID), e.Balance, e.Amount)
}

type AccountNotTransactableError struct {
	AccountID AccountID
	Status    AccountStatus
}

func (e *AccountNotTransactableError) Error() string {
	return fmt.Sprintf("account %s with status %s is not allowed to transact", uuid.UUID(e.AccountID), e.Status)
}

type TransferNotAllowedError struct {
	UserType string
}

func (e *TransferNotAllowedError) Error() string {
	return fmt.Sprintf("%s users are not allowed to transfer", e.UserType)
}
