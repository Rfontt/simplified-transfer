package account

import (
	"time"

	"github.com/google/uuid"
)

type TransferCreatedEvent struct {
	ID        uuid.UUID
	CreatedAt time.Time
	Amount    float64
	To        AccountID
	From      AccountID
}

type TransferSucceeded struct {
	ID        uuid.UUID
	CreatedAt time.Time
	Amount    float64
	From      AccountID
}

type TransferFailed struct {
	ID        uuid.UUID
	CreatedAt time.Time
	Amount    float64
	To        AccountID
	From      AccountID
}

type DepositCreatedEvent struct {
	ID        uuid.UUID
	CreatedAt time.Time
	Amount    float64
	OwnerId   AccountID
}
