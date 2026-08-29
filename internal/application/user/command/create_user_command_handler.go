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
	name, err := user.NewFullName(cmd.FullName)
	if err != nil {
		return nil, mapUserCreationValidationError(err)
	}
	doc, err := user.NewDocument(cmd.Document)
	if err != nil {
		return nil, mapUserCreationValidationError(err)
	}
	email, err := user.NewEmail(cmd.Email)
	if err != nil {
		return nil, mapUserCreationValidationError(err)
	}
	userType, err := user.ParseType(cmd.Type)
	if err != nil {
		return nil, mapUserCreationValidationError(err)
	}
	plain, err := user.NewPassword(cmd.Password)
	if err != nil {
		return nil, mapUserCreationValidationError(err)
	}

	passwordHash, err := h.hasher.Hash(plain.String())
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	id := uuid.New()
	u, err := user.NewUser(user.ID(id), name, doc, email, passwordHash, userType)
	if err != nil {
		return nil, mapUserCreationValidationError(err)
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
		FullName: u.FullName.String(),
		Document: u.Document.String(),
		Email:    u.Email.String(),
		Type:     string(u.Type),
	}, nil
}

func mapUserCreationValidationError(err error) error {
	var (
		invalidFullName *user.InvalidFullNameError
		invalidDocument *user.InvalidDocumentError
		invalidEmail    *user.InvalidEmailError
		invalidPassword *user.InvalidPasswordError
		invalidType     *user.InvalidTypeError
	)
	switch {
	case errors.As(err, &invalidFullName):
		return fmt.Errorf("%w: %v", ErrInvalidFullName, err)
	case errors.As(err, &invalidDocument):
		return fmt.Errorf("%w: %v", ErrInvalidDocument, err)
	case errors.As(err, &invalidEmail):
		return fmt.Errorf("%w: %v", ErrInvalidEmail, err)
	case errors.As(err, &invalidPassword):
		return fmt.Errorf("%w: %v", ErrInvalidPassword, err)
	case errors.As(err, &invalidType):
		return fmt.Errorf("%w: %v", ErrInvalidType, err)
	default:
		return err
	}
}
