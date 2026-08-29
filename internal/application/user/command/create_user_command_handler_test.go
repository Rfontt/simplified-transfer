package command

import (
	"context"
	"errors"
	"testing"

	"event-driven-architecture/internal/domain"
	"event-driven-architecture/internal/domain/user"

	"github.com/google/uuid"
)

type fakeUserRepository struct {
	users  map[uuid.UUID]*user.User
	addErr error
}

func newFakeUserRepository() *fakeUserRepository {
	return &fakeUserRepository{users: make(map[uuid.UUID]*user.User)}
}

func (f *fakeUserRepository) Add(_ context.Context, u *user.User) error {
	if f.addErr != nil {
		return f.addErr
	}
	f.users[uuid.UUID(u.ID)] = u
	return nil
}

func (f *fakeUserRepository) ByID(_ context.Context, id user.ID) (*user.User, error) {
	return f.users[uuid.UUID(id)], nil
}

func (f *fakeUserRepository) AllUsers(_ context.Context) ([]*user.User, error) {
	out := make([]*user.User, 0, len(f.users))
	for _, u := range f.users {
		out = append(out, u)
	}
	return out, nil
}

type fakePasswordHasher struct {
	hash  string
	err   error
	calls int
}

func (f *fakePasswordHasher) Hash(_ string) (string, error) {
	f.calls++
	if f.err != nil {
		return "", f.err
	}
	return f.hash, nil
}

func validCreateUserCommand() CreateUserCommand {
	return CreateUserCommand{
		FullName: "Rita Fontenele",
		Document: "52998224725",
		Email:    "rita@example.com",
		Password: "secret",
		Type:     "common",
	}
}

func TestCreateUserCommandHandler_Success(t *testing.T) {
	repo := newFakeUserRepository()
	hasher := &fakePasswordHasher{hash: "$2a$10$hashed"}
	h := NewCreateUserCommandHandler(repo, hasher)

	result, err := h.Handle(context.Background(), validCreateUserCommand())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID == "" {
		t.Error("expected non-empty id")
	}
	if result.FullName != "Rita Fontenele" || result.Document != "52998224725" ||
		result.Email != "rita@example.com" || result.Type != "common" {
		t.Errorf("unexpected result: %+v", result)
	}
	if len(repo.users) != 1 {
		t.Fatalf("expected 1 saved user, got %d", len(repo.users))
	}
	var saved *user.User
	for _, u := range repo.users {
		saved = u
	}
	if saved.PasswordHash != "$2a$10$hashed" {
		t.Errorf("expected hashed password stored, got %q", saved.PasswordHash)
	}
	if saved.Type != user.COMMON {
		t.Errorf("expected COMMON type, got %v", saved.Type)
	}
	if hasher.calls != 1 {
		t.Errorf("expected 1 hash call, got %d", hasher.calls)
	}
}

func TestCreateUserCommandHandler_InvalidFullName(t *testing.T) {
	cmd := validCreateUserCommand()
	cmd.FullName = "  "
	h := NewCreateUserCommandHandler(newFakeUserRepository(), &fakePasswordHasher{})

	_, err := h.Handle(context.Background(), cmd)
	assertConstraintError(t, err, "full_name")
}

func TestCreateUserCommandHandler_InvalidDocument(t *testing.T) {
	cmd := validCreateUserCommand()
	cmd.Document = "123"
	h := NewCreateUserCommandHandler(newFakeUserRepository(), &fakePasswordHasher{})

	_, err := h.Handle(context.Background(), cmd)
	assertConstraintError(t, err, "document")
}

func TestCreateUserCommandHandler_InvalidEmail(t *testing.T) {
	cmd := validCreateUserCommand()
	cmd.Email = ""
	h := NewCreateUserCommandHandler(newFakeUserRepository(), &fakePasswordHasher{})

	_, err := h.Handle(context.Background(), cmd)
	assertConstraintError(t, err, "email")
}

func TestCreateUserCommandHandler_InvalidPassword(t *testing.T) {
	cmd := validCreateUserCommand()
	cmd.Password = ""
	hasher := &fakePasswordHasher{}
	h := NewCreateUserCommandHandler(newFakeUserRepository(), hasher)

	_, err := h.Handle(context.Background(), cmd)
	assertConstraintError(t, err, "password")
	if hasher.calls != 0 {
		t.Errorf("expected hasher not to be called, got %d calls", hasher.calls)
	}
}

func TestCreateUserCommandHandler_InvalidType(t *testing.T) {
	cmd := validCreateUserCommand()
	cmd.Type = "admin"
	h := NewCreateUserCommandHandler(newFakeUserRepository(), &fakePasswordHasher{})

	_, err := h.Handle(context.Background(), cmd)
	assertConstraintError(t, err, "type")
}

func TestCreateUserCommandHandler_HashError(t *testing.T) {
	repo := newFakeUserRepository()
	hasher := &fakePasswordHasher{err: errors.New("boom")}
	h := NewCreateUserCommandHandler(repo, hasher)

	_, err := h.Handle(context.Background(), validCreateUserCommand())
	if err == nil {
		t.Fatal("expected error")
	}
	if len(repo.users) != 0 {
		t.Errorf("expected no user persisted on hash failure, got %d", len(repo.users))
	}
}

func TestCreateUserCommandHandler_AlreadyExists(t *testing.T) {
	repo := newFakeUserRepository()
	repo.addErr = &user.AlreadyExistsError{Email: "rita@example.com"}
	h := NewCreateUserCommandHandler(repo, &fakePasswordHasher{hash: "$2a$10$hashed"})

	_, err := h.Handle(context.Background(), validCreateUserCommand())
	if !errors.Is(err, ErrUserAlreadyExists) {
		t.Fatalf("expected ErrUserAlreadyExists, got %v", err)
	}
}

var _ CreateUserUseCase = (*CreateUserCommandHandler)(nil)

var _ user.UserRepository = (*fakeUserRepository)(nil)

var _ user.PasswordHasher = (*fakePasswordHasher)(nil)

func assertConstraintError(t *testing.T, err error, field string) {
	t.Helper()
	var invalid *domain.ConstraintValidationError
	if !errors.As(err, &invalid) {
		t.Fatalf("expected ConstraintValidationError, got %v", err)
	}
	if invalid.Field != field {
		t.Errorf("expected field %q, got %q", field, invalid.Field)
	}
}
