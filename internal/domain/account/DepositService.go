package account

import "event-driven-architecture/internal/domain"

type DepositService interface {
	DepositFunds(accountId AccountID, amount domain.MonetaryAmount) (*Deposit, error)
}
