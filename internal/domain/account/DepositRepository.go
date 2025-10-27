package account

import (
	"github.com/google/uuid"
)

type DepositRepository interface {
	GetOne(id uuid.UUID) (Deposit, error)
	Create(deposit *Deposit) error
}
