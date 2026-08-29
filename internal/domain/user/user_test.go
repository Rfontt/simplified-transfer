package user

import (
	"errors"
	"testing"

	"event-driven-architecture/internal/domain"

	"github.com/google/uuid"
)

type fakeHasher struct {
	hash  string
	err   error
	calls int
}

func (f *fakeHasher) Hash(_ string) (string, error) {
	f.calls++
	if f.err != nil {
		return "", f.err
	}
	return f.hash, nil
}

func TestNewUser(t *testing.T) {
	id := uuid.New()
	hasher := &fakeHasher{hash: "hash"}
	u, err := NewUser(ID(id), hasher, "Rita Fontenele", "52998224725", "rita@example.com", "secret", COMMON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.ID != ID(id) || u.FullName != "Rita Fontenele" {
		t.Errorf("unexpected user: %+v", u)
	}
	if u.Document != "52998224725" || u.Email != "rita@example.com" {
		t.Errorf("unexpected user: %+v", u)
	}
	if u.PasswordHash != "hash" || u.Type != COMMON {
		t.Errorf("unexpected user: %+v", u)
	}
	if hasher.calls != 1 {
		t.Errorf("expected 1 hash call, got %d", hasher.calls)
	}
}

func TestNewUser_NormalizesFields(t *testing.T) {
	u, err := NewUser(ID(uuid.New()), &fakeHasher{hash: "hash"}, "  Rita Fontenele ", "529.982.247-25", "  rita@example.com ", "secret", "common")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.FullName != "Rita Fontenele" {
		t.Errorf("expected trimmed full name, got %q", u.FullName)
	}
	if u.Document != "52998224725" {
		t.Errorf("expected normalized document, got %q", u.Document)
	}
	if u.Email != "rita@example.com" {
		t.Errorf("expected trimmed email, got %q", u.Email)
	}
}

func TestNewUser_InvalidFullName(t *testing.T) {
	hasher := &fakeHasher{hash: "hash"}
	_, err := NewUser(ID(uuid.New()), hasher, "   ", "52998224725", "rita@example.com", "secret", COMMON)
	assertConstraintError(t, err, "full_name")
	assertHasherNotCalled(t, hasher)
}

func TestNewUser_InvalidDocument(t *testing.T) {
	hasher := &fakeHasher{hash: "hash"}
	_, err := NewUser(ID(uuid.New()), hasher, "Rita Fontenele", "123", "rita@example.com", "secret", COMMON)
	assertConstraintError(t, err, "document")
	assertHasherNotCalled(t, hasher)
}

func TestNewUser_InvalidEmail(t *testing.T) {
	hasher := &fakeHasher{hash: "hash"}
	_, err := NewUser(ID(uuid.New()), hasher, "Rita Fontenele", "52998224725", "", "secret", COMMON)
	assertConstraintError(t, err, "email")
	assertHasherNotCalled(t, hasher)
}

func TestNewUser_InvalidPassword(t *testing.T) {
	hasher := &fakeHasher{hash: "hash"}
	_, err := NewUser(ID(uuid.New()), hasher, "Rita Fontenele", "52998224725", "rita@example.com", "  ", COMMON)
	assertConstraintError(t, err, "password")
	assertHasherNotCalled(t, hasher)
}

func TestNewUser_InvalidType(t *testing.T) {
	hasher := &fakeHasher{hash: "hash"}
	_, err := NewUser(ID(uuid.New()), hasher, "Rita Fontenele", "52998224725", "rita@example.com", "secret", "admin")
	assertConstraintError(t, err, "type")
	assertHasherNotCalled(t, hasher)
}

func TestNewUser_EmptyHash(t *testing.T) {
	_, err := NewUser(ID(uuid.New()), &fakeHasher{}, "Rita Fontenele", "52998224725", "rita@example.com", "secret", COMMON)
	assertConstraintError(t, err, "password")
}

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

func assertHasherNotCalled(t *testing.T, hasher *fakeHasher) {
	t.Helper()
	if hasher.calls != 0 {
		t.Errorf("expected hasher not to be called, got %d calls", hasher.calls)
	}
}

func TestValidateDocument(t *testing.T) {
	u := User{}
	valid := []string{
		"52998224725",
		"529.982.247-25",
		"11222333000181",
		"11.222.333/0001-81",
	}
	for _, doc := range valid {
		u.Document = doc
		digits, err := u.validateDocument()
		if err != nil {
			t.Errorf("expected %s to be valid, got %v", doc, err)
		}
		if digits != onlyDigits(doc) {
			t.Errorf("expected normalized document %s, got %s", onlyDigits(doc), digits)
		}
	}

	invalid := []string{
		"",
		"123",
		"52998224724",
		"11222333000182",
		"11111111111",
		"11111111111111",
	}
	for _, doc := range invalid {
		u.Document = doc
		if _, err := u.validateDocument(); err == nil {
			t.Errorf("expected %s to be invalid", doc)
		}
	}
}

func TestUserCanTransfer(t *testing.T) {
	common := User{Type: COMMON}
	if !common.CanTransfer() {
		t.Error("expected COMMON user to be able to transfer")
	}
	shopkeeper := User{Type: SHOPKEEPER}
	if shopkeeper.CanTransfer() {
		t.Error("expected SHOPKEEPER user to not be able to transfer")
	}
}
