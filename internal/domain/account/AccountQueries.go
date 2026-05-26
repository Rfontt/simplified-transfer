package account

import (
	"context"

	"github.com/google/uuid"
)

type AccountQueries interface {
	OwnerBalance(ctx context.Context, ownerID uuid.UUID) (*Account, error)
}
