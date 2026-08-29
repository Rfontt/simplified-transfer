package command

import (
	"context"
	"errors"
	"testing"

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
	if saved.Password != "$2a$10$hashed" {
		t.Errorf("expected hashed password stored, got %q", saved.Password)
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
	if !errors.Is(err, ErrInvalidFullName) {
		t.Fatalf("expected ErrInvalidFullName, got %v", err)
	}
}

func TestCreateUserCommandHandler_InvalidDocument(t *testing.T) {
	cmd := validCreateUserCommand()
	cmd.Document = "123"
	h := NewCreateUserCommandHandler(newFakeUserRepository(), &fakePasswordHasher{})

	_, err := h.Handle(context.Background(), cmd)
	if !errors.Is(err, ErrInvalidDocument) {
		t.Fatalf("expected ErrInvalidDocument, got %v", err)
	}
}

func TestCreateUserCommandHandler_InvalidEmail(t *testing.T) {
	cmd := validCreateUserCommand()
	cmd.Email = ""
	h := NewCreateUserCommandHandler(newFakeUserRepository(), &fakePasswordHasher{})

	_, err := h.Handle(context.Background(), cmd)
	if !errors.Is(err, ErrInvalidEmail) {
		t.Fatalf("expected ErrInvalidEmail, got %v", err)
	}
}

func TestCreateUserCommandHandler_InvalidPassword(t *testing.T) {
	cmd := validCreateUserCommand()
	cmd.Password = ""
	h := NewCreateUserCommandHandler(newFakeUserRepository(), &fakePasswordHasher{})

	_, err := h.Handle(context.Background(), cmd)
	if !errors.Is(err, ErrInvalidPassword) {
		t.Fatalf("expected ErrInvalidPassword, got %v", err)
	}
}

func TestCreateUserCommandHandler_InvalidType(t *testing.T) {
	cmd := validCreateUserCommand()
	cmd.Type = "admin"
	h := NewCreateUserCommandHandler(newFakeUserRepository(), &fakePasswordHasher{})

	_, err := h.Handle(context.Background(), cmd)
	if !errors.Is(err, ErrInvalidType) {
		t.Fatalf("expected ErrInvalidType, got %v", err)
	}
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
	h := NewCreateUserCommandHandler(repo, &fakePasswordHasher{})

	_, err := h.Handle(context.Background(), validCreateUserCommand())
	if !errors.Is(err, ErrUserAlreadyExists) {
		t.Fatalf("expected ErrUserAlreadyExists, got %v", err)
	}
}

var _ CreateUserUseCase = (*CreateUserCommandHandler)(nil)

var _ user.UserRepository = (*fakeUserRepository)(nil)

var _ user.PasswordHasher = (*fakePasswordHasher)(nil)
