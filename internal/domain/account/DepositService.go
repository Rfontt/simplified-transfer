package account

import "event-driven-architecture/internal/domain"

type DepositService interface {
	Create(accountId AccountID, amount domain.MonetaryAmount) error
}
