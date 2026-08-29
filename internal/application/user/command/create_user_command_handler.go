package command

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"event-driven-architecture/internal/domain/user"

	"github.com/google/uuid"
)

type CreateUserUseCase interface {
	Handle(ctx context.Context, cmd CreateUserCommand) (*CreateUserResult, error)
}

type CreateUserResult struct {
	ID       string
	FullName string
	Document string
	Email    string
	Type     string
}

type CreateUserCommandHandler struct {
	users  user.UserRepository
	hasher user.PasswordHasher
}

func NewCreateUserCommandHandler(users user.UserRepository, hasher user.PasswordHasher) *CreateUserCommandHandler {
	return &CreateUserCommandHandler{users: users, hasher: hasher}
}

func (h *CreateUserCommandHandler) Handle(ctx context.Context, cmd CreateUserCommand) (*CreateUserResult, error) {
	if strings.TrimSpace(cmd.FullName) == "" {
		return nil, ErrInvalidFullName
	}

	if err := user.ValidateDocument(cmd.Document); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidDocument, err)
	}

	if cmd.Email == "" {
		return nil, ErrInvalidEmail
	}

	if cmd.Password == "" {
		return nil, ErrInvalidPassword
	}

	userType, err := user.ParseType(cmd.Type)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidType, err)
	}

	passwordHash, err := h.hasher.Hash(cmd.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	id := uuid.New()
	u := user.NewUser(user.ID(id), cmd.FullName, cmd.Document, cmd.Email, passwordHash, userType)

	if err := h.users.Add(ctx, u); err != nil {
		var alreadyExists *user.AlreadyExistsError
		if errors.As(err, &alreadyExists) {
			return nil, fmt.Errorf("%w: %v", ErrUserAlreadyExists, alreadyExists)
		}
		return nil, fmt.Errorf("failed to persist user: %w", err)
	}

	return &CreateUserResult{
		ID:       id.String(),
		FullName: u.FullName,
		Document: u.Document,
		Email:    u.Email,
		Type:     string(u.Type),
	}, nil
}
