package query

import (
	"context"

	"event-driven-architecture/internal/domain"
	"event-driven-architecture/internal/domain/account"
)

type UserBalanceQueryHandler struct {
	accountQueries account.AccountQueries
}

func NewUserBalanceQueryHandler(accountQueries account.AccountQueries) *UserBalanceQueryHandler {
	return &UserBalanceQueryHandler{
		accountQueries: accountQueries,
	}
}

func (h *UserBalanceQueryHandler) Handle(ctx context.Context, query UserBalanceQuery) (*UserBalanceQueryProjection, error) {
	userAggregateId, err := domain.ToAggregateID(query.UserID)

	if err != nil {
		return nil, err
	}

	acc, err := h.accountQueries.OwnerBalance(ctx, userAggregateId)

	if err != nil {
		return nil, err
	}

	return &UserBalanceQueryProjection{
		Balance:  acc.Balance,
		Currency: acc.Balance.Currency,
	}, nil
}
