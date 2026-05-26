package account

import "context"

type DepositRepository interface {
	ByID(ctx context.Context, id string) (*Deposit, error)
	Add(ctx context.Context, deposit *Deposit) error

	AllDeposits(ctx context.Context) ([]*Deposit, error)
}
