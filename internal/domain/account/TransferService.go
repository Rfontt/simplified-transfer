package account

import "event-driven-architecture/internal/domain"

type TransferService interface {
	TransferFunds(from, to AccountID, amount domain.MonetaryAmount) (*Transfer, error)
}
