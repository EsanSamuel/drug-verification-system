// Package password provides secure password hashing and comparison using bcrypt.
package password

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// DefaultCost is the bcrypt cost used for hashing.
const DefaultCost = 12

// Hash returns a bcrypt hash of the password.
func Hash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", fmt.Errorf("password: hash: %w", err)
	}
	return string(bytes), nil
}

// Compare checks whether the given password matches the hash.
// Returns nil on match, or an error on mismatch.
func Compare(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
