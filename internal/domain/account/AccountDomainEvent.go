package account

import (
	"event-driven-architecture/internal/domain/user"
	"time"

	"github.com/google/uuid"
)

type AccountCreated struct {
	ID        uuid.UUID
	CreatedAt time.Time
	OwnerId   user.ID
}

type AccountTransferCreated struct {
	ID        uuid.UUID
	CreatedAt time.Time
	Amount    float64
	To        AccountID
	From      AccountID
}

type AccountTransferSucceeded struct {
	ID        uuid.UUID
	CreatedAt time.Time
	Amount    float64
	From      AccountID
}

type AccountTransferFailed struct {
	ID           uuid.UUID
	CreatedAt    time.Time
	Amount       float64
	To           AccountID
	From         AccountID
	ErrorMessage string
}

type AccountDepositCreated struct {
	ID        uuid.UUID
	CreatedAt time.Time
	Amount    float64
	OwnerId   AccountID
}

type AccountDepositSucceeded struct {
	ID        uuid.UUID
	CreatedAt time.Time
	Amount    float64
	OwnerId   AccountID
}

type AccountDepositFailed struct {
	ID           uuid.UUID
	CreatedAt    time.Time
	Amount       float64
	OwnerId      AccountID
	ErrorMessage string
}
