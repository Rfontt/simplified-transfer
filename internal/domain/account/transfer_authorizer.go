package account

import "context"

type TransferAuthorizer interface {
	Authorize(ctx context.Context, transferID string) error
}
