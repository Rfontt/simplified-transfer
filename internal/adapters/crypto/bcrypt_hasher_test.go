package crypto

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestBcryptHasher_Hash(t *testing.T) {
	h := NewBcryptHasher()

	hash, err := h.Hash("secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("secret")); err != nil {
		t.Errorf("hash does not verify: %v", err)
	}

	other, err := h.Hash("secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hash == other {
		t.Error("expected distinct hashes for the same password")
	}
}
