package password_test

import (
	"testing"

	"github.com/drug-verification/server/pkg/password"
)

func TestHashAndCompare(t *testing.T) {
	plain := "SecretP@ssw0rd!"

	hash, err := password.Hash(plain)
	if err != nil {
		t.Fatalf("unexpected error hashing password: %v", err)
	}

	if hash == "" || hash == plain {
		t.Fatalf("hash should not be empty or plain text")
	}

	// Correct password comparison
	if err := password.Compare(hash, plain); err != nil {
		t.Errorf("expected password to match hash, got error: %v", err)
	}

	// Incorrect password comparison
	if err := password.Compare(hash, "WrongPassword"); err == nil {
		t.Errorf("expected error for mismatched password, got nil")
	}
}

func TestDifferentHashesForSamePassword(t *testing.T) {
	plain := "MySecurePassword123"

	hash1, err := password.Hash(plain)
	if err != nil {
		t.Fatalf("hash1 failed: %v", err)
	}

	hash2, err := password.Hash(plain)
	if err != nil {
		t.Fatalf("hash2 failed: %v", err)
	}

	// bcrypt produces unique salts each time
	if hash1 == hash2 {
		t.Errorf("expected unique hashes due to salts, got identical hashes")
	}
}
