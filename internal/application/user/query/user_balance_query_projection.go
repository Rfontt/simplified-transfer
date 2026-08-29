package query

import (
	"event-driven-architecture/internal/domain"
)

type UserBalanceQueryProjection struct {
	UserID   string
	Balance  domain.MonetaryAmount
	Currency string
}
