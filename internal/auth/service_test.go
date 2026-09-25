package auth_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/drug-verification/server/internal/auth"
	"github.com/drug-verification/server/internal/database/sqlc"
	jwtpkg "github.com/drug-verification/server/pkg/jwt"
	"github.com/drug-verification/server/pkg/password"
)

type mockAuthRepo struct {
	usersByEmail map[string]sqlc.User
	usersByID    map[uuid.UUID]sqlc.GetUserByIDRow
}

func newMockAuthRepo() *mockAuthRepo {
	return &mockAuthRepo{
		usersByEmail: make(map[string]sqlc.User),
		usersByID:    make(map[uuid.UUID]sqlc.GetUserByIDRow),
	}
}

func (m *mockAuthRepo) CreateUser(ctx context.Context, params sqlc.CreateUserParams) (sqlc.CreateUserRow, error) {
	id := uuid.New()
	now := time.Now()
	user := sqlc.User{
		ID:           id,
		FullName:     params.FullName,
		Email:        params.Email,
		PasswordHash: params.PasswordHash,
		Role:         params.Role,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	m.usersByEmail[params.Email] = user
	m.usersByID[id] = sqlc.GetUserByIDRow{
		ID:        id,
		FullName:  params.FullName,
		Email:     params.Email,
		Role:      params.Role,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return sqlc.CreateUserRow{
		ID:        id,
		FullName:  params.FullName,
		Email:     params.Email,
		Role:      params.Role,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (m *mockAuthRepo) GetUserByEmail(ctx context.Context, email string) (sqlc.User, error) {
	u, ok := m.usersByEmail[email]
	if !ok {
		return sqlc.User{}, errors.New("not found")
	}
	return u, nil
}

func (m *mockAuthRepo) GetUserByID(ctx context.Context, id uuid.UUID) (sqlc.GetUserByIDRow, error) {
	u, ok := m.usersByID[id]
	if !ok {
		return sqlc.GetUserByIDRow{}, errors.New("not found")
	}
	return u, nil
}

func TestRegisterAndLogin(t *testing.T) {
	repo := newMockAuthRepo()
	jwtMgr := jwtpkg.NewManager("test-secret-key-12345678", 1*time.Hour)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := auth.NewService(repo, jwtMgr, logger)

	// Test Registration
	regReq := auth.RegisterRequest{
		FullName: "Pharmacist John",
		Email:    "john@pharmacy.com",
		Password: "SecurePassword123!",
	}

	resp, err := service.Register(context.Background(), regReq)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	if resp.Token == "" {
		t.Errorf("expected JWT token in response, got empty")
	}
	if resp.User.Email != "john@pharmacy.com" {
		t.Errorf("expected email john@pharmacy.com, got %q", resp.User.Email)
	}

	// Test Duplicate Email Registration
	_, err = service.Register(context.Background(), regReq)
	if !errors.Is(err, auth.ErrEmailTaken) {
		t.Errorf("expected ErrEmailTaken for duplicate email, got: %v", err)
	}

	// Test Login Success
	loginResp, err := service.Login(context.Background(), auth.LoginRequest{
		Email:    "john@pharmacy.com",
		Password: "SecurePassword123!",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if loginResp.Token == "" {
		t.Errorf("expected token in login response")
	}

	// Test Login Wrong Password
	_, err = service.Login(context.Background(), auth.LoginRequest{
		Email:    "john@pharmacy.com",
		Password: "WrongPassword!",
	})
	if !errors.Is(err, auth.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials for wrong password, got: %v", err)
	}
}

func TestRegister_Validation(t *testing.T) {
	repo := newMockAuthRepo()
	jwtMgr := jwtpkg.NewManager("test-secret-key-12345678", 1*time.Hour)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := auth.NewService(repo, jwtMgr, logger)

	// Missing full name
	_, err := service.Register(context.Background(), auth.RegisterRequest{
		FullName: "",
		Email:    "test@test.com",
		Password: "Password123!",
	})
	if err == nil {
		t.Errorf("expected error for empty full name")
	}

	// Short password
	_, err = service.Register(context.Background(), auth.RegisterRequest{
		FullName: "Test User",
		Email:    "test@test.com",
		Password: "short",
	})
	if err == nil {
		t.Errorf("expected error for password < 8 chars")
	}
}

func TestGetCurrentUser(t *testing.T) {
	repo := newMockAuthRepo()
	jwtMgr := jwtpkg.NewManager("test-secret-key-12345678", 1*time.Hour)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := auth.NewService(repo, jwtMgr, logger)

	pwdHash, _ := password.Hash("Password123!")
	created, _ := repo.CreateUser(context.Background(), sqlc.CreateUserParams{
		FullName:     "Alice Smith",
		Email:        "alice@pharmacy.com",
		PasswordHash: pwdHash,
		Role:         sqlc.UserRolePharmacist,
	})

	user, err := service.GetCurrentUser(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("get current user failed: %v", err)
	}

	if user.ID != created.ID {
		t.Errorf("expected user id %v, got %v", created.ID, user.ID)
	}
	if user.Email != "alice@pharmacy.com" {
		t.Errorf("expected alice@pharmacy.com, got %s", user.Email)
	}
}
