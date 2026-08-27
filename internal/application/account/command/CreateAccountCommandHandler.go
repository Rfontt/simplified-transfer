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

// CreateAccountUseCase is the driving port consumed by adapters (HTTP) to
// create accounts without depending on the concrete handler.
type CreateAccountUseCase interface {
	Handle(ctx context.Context, cmd CreateAccountCommand) (*CreateAccountResult, error)
}

// CreateAccountResult is the transport-agnostic result returned by the use
// case, so the domain Account entity does not leak into adapters.
type CreateAccountResult struct {
	ID       string
	OwnerID  string
	Currency string
	Balance  float64
}

// CreateAccountCommandHandler implements CreateAccountUseCase.
type CreateAccountCommandHandler struct {
	accounts account.AccountRepository
}

func NewCreateAccountCommandHandler(accounts account.AccountRepository) *CreateAccountCommandHandler {
	return &CreateAccountCommandHandler{accounts: accounts}
}

func (h *CreateAccountCommandHandler) Handle(ctx context.Context, cmd CreateAccountCommand) (*CreateAccountResult, error) {
	ownerID, err := uuid.Parse(cmd.OwnerID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidOwnerID, err)
	}

	if cmd.Currency == "" {
		return nil, ErrInvalidCurrency
	}

	if cmd.Balance < 0 {
		return nil, ErrInvalidBalance
	}

	id := uuid.New()
	acc := account.NewAccount(
		account.AccountID(id),
		user.ID(ownerID),
		domain.MonetaryAmount{Currency: cmd.Currency, Value: cmd.Balance},
	)

	if err := h.accounts.Add(ctx, acc); err != nil {
		var alreadyExists *account.AccountAlreadyExistsError
		var ownerNotFound *account.OwnerNotFoundError
		switch {
		case errors.As(err, &alreadyExists):
			return nil, fmt.Errorf("%w: %v", ErrAccountAlreadyExists, alreadyExists)
		case errors.As(err, &ownerNotFound):
			return nil, fmt.Errorf("%w: %v", ErrOwnerNotFound, ownerNotFound)
		default:
			return nil, fmt.Errorf("failed to persist account: %w", err)
		}
	}

	return &CreateAccountResult{
		ID:       id.String(),
		OwnerID:  ownerID.String(),
		Currency: cmd.Currency,
		Balance:  cmd.Balance,
	}, nil
}
