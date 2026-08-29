package user

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewUser(t *testing.T) {
	id := uuid.New()
	u, err := NewUser(ID(id), FullName("Rita Fontenele"), Document("52998224725"), Email("rita@example.com"), "hash", COMMON)
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
}

func TestNewUser_EmptyPasswordHash(t *testing.T) {
	if _, err := NewUser(ID(uuid.New()), FullName("Rita Fontenele"), Document("52998224725"), Email("rita@example.com"), "", COMMON); err == nil {
		t.Error("expected error for empty password hash")
	}
}

func TestNewDocument(t *testing.T) {
	valid := []string{
		"52998224725",
		"529.982.247-25",
		"11222333000181",
		"11.222.333/0001-81",
	}
	for _, doc := range valid {
		d, err := NewDocument(doc)
		if err != nil {
			t.Errorf("expected %s to be valid, got %v", doc, err)
		}
		if string(d) != onlyDigits(doc) {
			t.Errorf("expected normalized document %s, got %s", onlyDigits(doc), string(d))
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
		if _, err := NewDocument(doc); err == nil {
			t.Errorf("expected %s to be invalid", doc)
		}
	}
}

func TestNewFullName(t *testing.T) {
	name, err := NewFullName("  Rita Fontenele ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "Rita Fontenele" {
		t.Errorf("expected trimmed name, got %q", name)
	}

	for _, v := range []string{"", "   "} {
		if _, err := NewFullName(v); err == nil {
			t.Errorf("expected error for %q", v)
		}
	}
}

func TestNewEmail(t *testing.T) {
	email, err := NewEmail("  rita@example.com ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if email != "rita@example.com" {
		t.Errorf("expected trimmed email, got %q", email)
	}

	if _, err := NewEmail(""); err == nil {
		t.Error("expected error for empty email")
	}
}

func TestNewPassword(t *testing.T) {
	p, err := NewPassword("secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.String() != "secret" {
		t.Errorf("expected secret, got %q", p.String())
	}

	withSpaces, err := NewPassword(" secret ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if withSpaces.String() != " secret " {
		t.Errorf("expected untrimmed password, got %q", withSpaces.String())
	}

	for _, v := range []string{"", "   "} {
		if _, err := NewPassword(v); err == nil {
			t.Errorf("expected error for %q", v)
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

func TestParseType(t *testing.T) {
	for _, valid := range []string{"common", "shopkeeper"} {
		parsed, err := ParseType(valid)
		if err != nil {
			t.Errorf("expected %s to be valid, got %v", valid, err)
		}
		if string(parsed) != valid {
			t.Errorf("expected %s, got %s", valid, parsed)
		}
	}

	if _, err := ParseType("admin"); err == nil {
		t.Error("expected error for invalid type")
	}
}
