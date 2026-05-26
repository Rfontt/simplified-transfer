package account

import (
	"context"
)

type TransferRepository interface {
	ByID(ctx context.Context, id string) (*Transfer, error)
	Add(ctx context.Context, transfer *Transfer) error
}
