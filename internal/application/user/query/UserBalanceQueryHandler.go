package query

import (
	"event-driven-architecture/internal/domain"
	"event-driven-architecture/internal/domain/account"
)

type UserBalanceQueryHandler struct {
	accountService account.AccountService
}

func NewUserBalanceQueryHandler(accountService account.AccountService) *UserBalanceQueryHandler {
	return &UserBalanceQueryHandler{
		accountService: accountService,
	}
}

func (h *UserBalanceQueryHandler) Handle(query UserBalanceQuery) (*UserBalanceQueryProjection, error) {
	userAggregateId, err := domain.ToAggregateID(query.UserID)

	if err != nil {
		return nil, err
	}

	acc, err := h.accountService.GetOwnerBalance(userAggregateId)

	if err != nil {
		return nil, err
	}

	return &UserBalanceQueryProjection{
		Balance:  acc.Balance,
		Currency: acc.Balance.Currency,
	}, nil
}
