package jwt_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/drug-verification/server/pkg/jwt"
)

func TestGenerateAndValidate(t *testing.T) {
	secret := "super-secure-secret-key-12345"
	mgr := jwt.NewManager(secret, 15*time.Minute)

	userID := uuid.New()
	role := "pharmacist"

	tokenStr, err := mgr.Generate(userID, role)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	claims, err := mgr.Validate(tokenStr)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("expected userID %v, got %v", userID, claims.UserID)
	}
	if claims.Role != role {
		t.Errorf("expected role %v, got %v", role, claims.Role)
	}
}

func TestExpiredToken(t *testing.T) {
	secret := "super-secure-secret-key-12345"
	// 1 millisecond expiration
	mgr := jwt.NewManager(secret, -1*time.Minute)

	userID := uuid.New()
	tokenStr, err := mgr.Generate(userID, "admin")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	_, err = mgr.Validate(tokenStr)
	if !errors.Is(err, jwt.ErrTokenExpired) {
		t.Errorf("expected ErrTokenExpired, got: %v", err)
	}
}

func TestInvalidSecret(t *testing.T) {
	mgr1 := jwt.NewManager("secret-key-1", 15*time.Minute)
	mgr2 := jwt.NewManager("secret-key-2", 15*time.Minute)

	userID := uuid.New()
	tokenStr, err := mgr1.Generate(userID, "pharmacist")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	_, err = mgr2.Validate(tokenStr)
	if !errors.Is(err, jwt.ErrTokenInvalid) {
		t.Errorf("expected ErrTokenInvalid when validating with wrong secret, got: %v", err)
	}
}
