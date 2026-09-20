package command

import (
	"context"
	"errors"
	"fmt"

	"event-driven-architecture/internal/domain"
	"event-driven-architecture/internal/domain/account"
	"event-driven-architecture/internal/domain/user"

	"github.com/google/uuid"
)

type TransferMoneyUseCase interface {
	Handle(ctx context.Context, cmd TransferMoneyCommand) (*TransferMoneyResult, error)
}

type TransferMoneyResult struct {
	TransferID     string
	Status         string
	Value          float64
	Currency       string
	PayerAccountID string
	PayeeAccountID string
	Events         []domain.Event
}

type TransferMoneyCommandHandler struct {
	accounts   account.AccountRepository
	users      user.UserRepository
	transfers  account.TransferRepository
	authorizer account.TransferAuthorizer
	publisher  TransferEventPublisher
}

func NewTransferMoneyCommandHandler(
	accounts account.AccountRepository,
	users user.UserRepository,
	transfers account.TransferRepository,
	authorizer account.TransferAuthorizer,
	publisher TransferEventPublisher,
) *TransferMoneyCommandHandler {
	return &TransferMoneyCommandHandler{
		accounts:   accounts,
		users:      users,
		transfers:  transfers,
		authorizer: authorizer,
		publisher:  publisher,
	}
}

func (h *TransferMoneyCommandHandler) Handle(ctx context.Context, cmd TransferMoneyCommand) (*TransferMoneyResult, error) {
	payerID, err := uuid.Parse(cmd.PayerAccountID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPayerAccountID, err)
	}
	payeeID, err := uuid.Parse(cmd.PayeeAccountID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPayeeAccountID, err)
	}

	from, err := h.loadAccount(ctx, payerID, ErrPayerAccountNotFound)
	if err != nil {
		return nil, err
	}
	to, err := h.loadAccount(ctx, payeeID, ErrPayeeAccountNotFound)
	if err != nil {
		return nil, err
	}

	payer, err := h.users.ByID(ctx, from.OwnerId)
	if err != nil {
		return nil, fmt.Errorf("failed to load payer: %w", err)
	}

	transferID := uuid.New()
	amount := domain.MonetaryAmount{Currency: from.Balance.Currency, Value: cmd.Value}

	if err := h.authorizer.Authorize(ctx, transferID.String()); err != nil {
		transfer := account.NewFailedTransfer(transferID, from.ID, to.ID, amount, err.Error())
		if saveErr := h.transfers.Save(ctx, transfer, from, to); saveErr != nil {
			return nil, fmt.Errorf("failed to persist transfer: %w", saveErr)
		}
		if publishErr := h.publish(ctx, transfer); publishErr != nil {
			return nil, publishErr
		}
		return nil, fmt.Errorf("%w: %v", ErrTransferNotAuthorized, err)
	}

	transfer, domainErr := account.NewTransfer(transferID, from, to, payer, amount)
	if err := h.transfers.Save(ctx, transfer, from, to); err != nil {
		return nil, fmt.Errorf("failed to persist transfer: %w", err)
	}
	if err := h.publish(ctx, transfer); err != nil {
		return nil, err
	}
	if domainErr != nil {
		return nil, mapTransferError(domainErr)
	}

	return &TransferMoneyResult{
		TransferID:     transferID.String(),
		Status:         string(transfer.Status),
		Value:          transfer.Amount.Value,
		Currency:       transfer.Amount.Currency,
		PayerAccountID: uuid.UUID(from.ID).String(),
		PayeeAccountID: uuid.UUID(to.ID).String(),
		Events:         transfer.UncommittedEvents(),
	}, nil
}

func (h *TransferMoneyCommandHandler) loadAccount(ctx context.Context, id uuid.UUID, notFound error) (*account.Account, error) {
	acc, err := h.accounts.ByID(ctx, account.AccountID(id))
	if err != nil {
		var accountNotFound *account.AccountNotFoundError
		if errors.As(err, &accountNotFound) {
			return nil, fmt.Errorf("%w: %v", notFound, accountNotFound)
		}
		return nil, fmt.Errorf("failed to load account: %w", err)
	}
	return acc, nil
}

func (h *TransferMoneyCommandHandler) publish(ctx context.Context, transfer *account.Transfer) error {
	for _, event := range transfer.UncommittedEvents() {
		if err := h.publisher.Publish(ctx, event); err != nil {
			return fmt.Errorf("%w: %v", ErrTransferEventPublishFailed, err)
		}
	}
	return nil
}

func mapTransferError(err error) error {
	var sameAccount *account.SameAccountTransferError
	var currencyMismatch *account.CurrencyMismatchError
	var insufficientBalance *account.InsufficientBalanceError
	var notTransactable *account.AccountNotTransactableError
	var notAllowed *account.TransferNotAllowedError

	switch {
	case errors.As(err, &sameAccount):
		return fmt.Errorf("%w: %v", ErrSameAccountTransfer, sameAccount)
	case errors.As(err, &currencyMismatch):
		return fmt.Errorf("%w: %v", ErrCurrencyMismatch, currencyMismatch)
	case errors.As(err, &insufficientBalance):
		return fmt.Errorf("%w: %v", ErrInsufficientBalance, insufficientBalance)
	case errors.As(err, &notTransactable):
		return fmt.Errorf("%w: %v", ErrAccountNotTransactable, notTransactable)
	case errors.As(err, &notAllowed):
		return fmt.Errorf("%w: %v", ErrTransferNotAllowed, notAllowed)
	default:
		return err
	}
}
