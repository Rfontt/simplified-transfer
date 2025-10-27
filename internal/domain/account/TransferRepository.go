package account

import (
	"github.com/google/uuid"
)

type TransferRepository interface {
	GetOne(id uuid.UUID) (Transfer, error)
	Save(transfer *Transfer) error
}
