package user

import (
	"time"

	"github.com/google/uuid"
)

type UserCreated struct {
	ID        uuid.UUID
	CreatedAt time.Time
	OwnerId   ID
}
