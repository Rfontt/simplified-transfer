package account

import "context"

type TransferRepository interface {
	ByID(ctx context.Context, id string) (*Transfer, error)
	Save(ctx context.Context, transfer *Transfer, from, to *Account) error

	AllTransfers(ctx context.Context) ([]*Transfer, error)
}
