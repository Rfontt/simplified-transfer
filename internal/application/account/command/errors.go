package command

import "errors"

var (
	ErrAccountAlreadyExists = errors.New("account already exists for this owner")
	ErrOwnerNotFound        = errors.New("owner not found")
	ErrInvalidOwnerID       = errors.New("owner id must be a valid uuid")

	ErrInvalidPayerAccountID      = errors.New("payer account id must be a valid uuid")
	ErrInvalidPayeeAccountID      = errors.New("payee account id must be a valid uuid")
	ErrPayerAccountNotFound       = errors.New("payer account not found")
	ErrPayeeAccountNotFound       = errors.New("payee account not found")
	ErrTransferNotAuthorized      = errors.New("transfer not authorized")
	ErrInsufficientBalance        = errors.New("insufficient balance")
	ErrAccountNotTransactable     = errors.New("payer account is not allowed to transact")
	ErrTransferNotAllowed         = errors.New("payer is not allowed to transfer")
	ErrSameAccountTransfer        = errors.New("payer and payee accounts must differ")
	ErrCurrencyMismatch           = errors.New("accounts must hold the same currency")
	ErrTransferEventPublishFailed = errors.New("failed to publish transfer events")
)
