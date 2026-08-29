package domain

import "github.com/google/uuid"

type AggregateID uuid.UUID

func ToAggregateID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
