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

func NewUser(id ID, fullName, document, email, password string, userType Type) *User {
	return &User{
		ID:       id,
		FullName: fullName,
		Document: document,
		Email:    email,
		Password: password,
		Type:     userType,
	}
}

func ParseType(value string) (Type, error) {
	t := Type(value)
	switch t {
	case COMMON, SHOPKEEPER:
		return t, nil
	default:
		return "", &InvalidTypeError{Type: value}
	}
}

func (u User) CanTransfer() bool {
	return u.Type == COMMON
}
