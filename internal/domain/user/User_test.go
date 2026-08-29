package user

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewUser(t *testing.T) {
	id := uuid.New()
	u := NewUser(ID(id), "Rita Fontenele", "52998224725", "rita@example.com", "hash", COMMON)
	if u.ID != ID(id) || u.FullName != "Rita Fontenele" {
		t.Errorf("unexpected user: %+v", u)
	}
	if u.Document != "52998224725" || u.Email != "rita@example.com" {
		t.Errorf("unexpected user: %+v", u)
	}
	if u.Password != "hash" || u.Type != COMMON {
		t.Errorf("unexpected user: %+v", u)
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

func TestValidateDocument(t *testing.T) {
	valid := []string{
		"52998224725",
		"529.982.247-25",
		"11222333000181",
		"11.222.333/0001-81",
	}
	for _, doc := range valid {
		if err := ValidateDocument(doc); err != nil {
			t.Errorf("expected %s to be valid, got %v", doc, err)
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
		if err := ValidateDocument(doc); err == nil {
			t.Errorf("expected %s to be invalid", doc)
		}
	}
}
