package command

import (
	"context"

	"event-driven-architecture/internal/domain"
)

type TransferEventPublisher interface {
	Publish(ctx context.Context, event domain.Event) error
}
