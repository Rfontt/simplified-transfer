package command

import (
	"context"
	"errors"
	"fmt"

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
	id := uuid.New()
	u, err := user.NewUser(user.ID(id), h.hasher, cmd.FullName, cmd.Document, cmd.Email, cmd.Password, cmd.Type)
	if err != nil {
		return nil, err
	}

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
		Type:     u.Type,
	}, nil
}
