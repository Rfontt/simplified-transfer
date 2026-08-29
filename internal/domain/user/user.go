package user

import "event-driven-architecture/internal/domain"

type Type string

const (
	COMMON     Type = "common"
	SHOPKEEPER Type = "shopkeeper"
)

type ID domain.AggregateID

type User struct {
	ID           ID
	FullName     FullName
	Document     Document
	Email        Email
	PasswordHash string
	Type         Type
}

func NewUser(id ID, fullName FullName, document Document, email Email, passwordHash string, userType Type) (*User, error) {
	if passwordHash == "" {
		return nil, &InvalidPasswordError{}
	}
	return &User{
		ID:           id,
		FullName:     fullName,
		Document:     document,
		Email:        email,
		PasswordHash: passwordHash,
		Type:         userType,
	}, nil
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
