package user

import "event-driven-architecture/internal/domain"

type Type string

const (
	COMMON     Type = "common"
	SHOPKEEPER Type = "shopkeeper"
)

type ID domain.AggregateID

type User struct {
	ID       ID
	FullName string
	Document string
	Email    string
	Password string
	Type     Type
}
