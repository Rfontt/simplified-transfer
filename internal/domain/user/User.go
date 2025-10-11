package user

import "event-driven-architecture/internal/domain"

type Type string

const (
	COMMON     Type = "common"
	SHOPKEEPER Type = "shopkeeper"
)

type User struct {
	ID       string
	Name     string
	LastName string
	Document string
	Email    string
	Password string
	Balance  domain.Balance
	Type     Type
}
