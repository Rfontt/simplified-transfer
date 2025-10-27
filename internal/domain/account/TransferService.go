package account

import "event-driven-architecture/internal/domain"

type TransferService interface {
	Create(from, to AccountID, amount domain.MonetaryAmount) error
}
