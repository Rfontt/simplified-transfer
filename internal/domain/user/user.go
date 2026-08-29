package user

import (
	"fmt"
	"strings"

	"event-driven-architecture/internal/domain"

	"github.com/paemuri/brdoc"
)

const (
	COMMON     = "common"
	SHOPKEEPER = "shopkeeper"
)

type ID domain.AggregateID

type User struct {
	ID           ID
	FullName     string
	Document     string
	Email        string
	PasswordHash string
	Type         string
}

func NewUser(id ID, hasher PasswordHasher, fullName, document, email, plainPassword, userType string) (*User, error) {
	u := &User{
		ID:       id,
		FullName: fullName,
		Document: document,
		Email:    email,
		Type:     userType,
	}
	if err := u.validateFields(hasher, plainPassword); err != nil {
		return nil, err
	}
	return u, nil
}

func (u User) CanTransfer() bool {
	return u.Type == COMMON
}

func (u *User) validateFields(hasher PasswordHasher, plainPassword string) error {
	u.FullName = strings.TrimSpace(u.FullName)
	if u.FullName == "" {
		return &domain.ConstraintValidationError{Field: "full_name"}
	}

	u.Email = strings.TrimSpace(u.Email)
	if u.Email == "" {
		return &domain.ConstraintValidationError{Field: "email"}
	}

	if strings.TrimSpace(plainPassword) == "" {
		return &domain.ConstraintValidationError{Field: "password"}
	}

	doc, err := u.validateDocument()
	if err != nil {
		return err
	}
	u.Document = doc

	switch u.Type {
	case COMMON, SHOPKEEPER:
	default:
		return &domain.ConstraintValidationError{Field: "type"}
	}

	return u.hashPassword(hasher, plainPassword)
}

func (u *User) hashPassword(hasher PasswordHasher, plainPassword string) error {
	hash, err := hasher.Hash(plainPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	if hash == "" {
		return &domain.ConstraintValidationError{Field: "password"}
	}
	u.PasswordHash = hash
	return nil
}

func (u User) validateDocument() (string, error) {
	digits := onlyDigits(u.Document)
	switch len(digits) {
	case 11:
		if brdoc.IsCPF(digits) {
			return digits, nil
		}
	case 14:
		if brdoc.IsCNPJ(digits) {
			return digits, nil
		}
	}
	return "", &domain.ConstraintValidationError{Field: "document"}
}

func onlyDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
