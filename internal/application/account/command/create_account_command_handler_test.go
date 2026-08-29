package command

import (
	"context"
	"errors"
	"testing"

	"event-driven-architecture/internal/domain"
	"event-driven-architecture/internal/domain/account"

	"github.com/google/uuid"
)

type fakeAccountRepository struct {
	accounts map[uuid.UUID]*account.Account
	addErr   error
}

func newFakeAccountRepository() *fakeAccountRepository {
	return &fakeAccountRepository{accounts: make(map[uuid.UUID]*account.Account)}
}

func (f *fakeAccountRepository) Add(_ context.Context, acc *account.Account) error {
	if f.addErr != nil {
		return f.addErr
	}
	f.accounts[uuid.UUID(acc.ID)] = acc
	return nil
}

func (f *fakeAccountRepository) ByID(_ context.Context, id account.AccountID) (*account.Account, error) {
	return f.accounts[uuid.UUID(id)], nil
}

func (f *fakeAccountRepository) AllAccounts(_ context.Context) ([]*account.Account, error) {
	out := make([]*account.Account, 0, len(f.accounts))
	for _, acc := range f.accounts {
		out = append(out, acc)
	}
	return out, nil
}

func TestCreateAccountCommandHandler_Success(t *testing.T) {
	repo := newFakeAccountRepository()
	h := NewCreateAccountCommandHandler(repo)

	ownerID := uuid.New().String()
	result, err := h.Handle(context.Background(), CreateAccountCommand{
		OwnerID:  ownerID,
		Currency: "BRL",
		Balance:  100,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID == "" {
		t.Error("expected non-empty id")
	}
	if result.OwnerID != ownerID {
		t.Errorf("expected owner id %s, got %s", ownerID, result.OwnerID)
	}
	if result.Currency != "BRL" || result.Balance != 100 {
		t.Errorf("unexpected result: %+v", result)
	}
	if len(repo.accounts) != 1 {
		t.Errorf("expected 1 saved account, got %d", len(repo.accounts))
	}
}

func TestCreateAccountCommandHandler_InvalidOwnerID(t *testing.T) {
	h := NewCreateAccountCommandHandler(newFakeAccountRepository())
	_, err := h.Handle(context.Background(), CreateAccountCommand{
		OwnerID:  "not-a-uuid",
		Currency: "BRL",
		Balance:  0,
	})
	if !errors.Is(err, ErrInvalidOwnerID) {
		t.Fatalf("expected ErrInvalidOwnerID, got %v", err)
	}
}

func TestCreateAccountCommandHandler_InvalidCurrency(t *testing.T) {
	h := NewCreateAccountCommandHandler(newFakeAccountRepository())
	_, err := h.Handle(context.Background(), CreateAccountCommand{
		OwnerID:  uuid.New().String(),
		Currency: "",
		Balance:  0,
	})
	assertConstraintError(t, err, "currency")
}

func TestCreateAccountCommandHandler_InvalidBalance(t *testing.T) {
	h := NewCreateAccountCommandHandler(newFakeAccountRepository())
	_, err := h.Handle(context.Background(), CreateAccountCommand{
		OwnerID:  uuid.New().String(),
		Currency: "BRL",
		Balance:  -1,
	})
	assertConstraintError(t, err, "balance")
}

func TestCreateAccountCommandHandler_AlreadyExists(t *testing.T) {
	repo := newFakeAccountRepository()
	repo.addErr = &account.AccountAlreadyExistsError{OwnerID: "x"}
	h := NewCreateAccountCommandHandler(repo)

	_, err := h.Handle(context.Background(), CreateAccountCommand{
		OwnerID:  uuid.New().String(),
		Currency: "BRL",
		Balance:  0,
	})
	if !errors.Is(err, ErrAccountAlreadyExists) {
		t.Fatalf("expected ErrAccountAlreadyExists, got %v", err)
	}
}

func TestCreateAccountCommandHandler_OwnerNotFound(t *testing.T) {
	repo := newFakeAccountRepository()
	repo.addErr = &account.OwnerNotFoundError{OwnerID: "x"}
	h := NewCreateAccountCommandHandler(repo)

	_, err := h.Handle(context.Background(), CreateAccountCommand{
		OwnerID:  uuid.New().String(),
		Currency: "BRL",
		Balance:  0,
	})
	if !errors.Is(err, ErrOwnerNotFound) {
		t.Fatalf("expected ErrOwnerNotFound, got %v", err)
	}
}

var _ CreateAccountUseCase = (*CreateAccountCommandHandler)(nil)

var _ account.AccountRepository = (*fakeAccountRepository)(nil)

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
