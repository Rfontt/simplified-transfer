package account

import "context"

type AccountRepository interface {
	ByID(ctx context.Context, id AccountID) (*Account, error)
	Add(ctx context.Context, account *Account) error

	AllAccounts(ctx context.Context) ([]*Account, error)
}
