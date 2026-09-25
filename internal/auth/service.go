package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/mail"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/drug-verification/server/internal/database/sqlc"
	jwtpkg "github.com/drug-verification/server/pkg/jwt"
	"github.com/drug-verification/server/pkg/password"
)

// Repository defines the data access interface for authentication.
type Repository interface {
	CreateUser(ctx context.Context, params sqlc.CreateUserParams) (sqlc.CreateUserRow, error)
	GetUserByEmail(ctx context.Context, email string) (sqlc.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (sqlc.GetUserByIDRow, error)
}

// Service implements authentication business logic.
type Service struct {
	repo       Repository
	jwtManager *jwtpkg.Manager
	logger     *slog.Logger
}

// NewService creates a new auth service.
func NewService(repo Repository, jwtManager *jwtpkg.Manager, logger *slog.Logger) *Service {
	return &Service{
		repo:       repo,
		jwtManager: jwtManager,
		logger:     logger,
	}
}

// RegisterRequest represents a user registration request.
type RegisterRequest struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest represents a user login request.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse is returned after successful authentication.
type AuthResponse struct {
	Token string   `json:"token"`
	User  UserInfo `json:"user"`
}

// UserInfo holds non-sensitive user data.
type UserInfo struct {
	ID       uuid.UUID `json:"id"`
	FullName string    `json:"full_name"`
	Email    string    `json:"email"`
	Role     string    `json:"role"`
}

// Sentinel errors.
var (
	ErrEmailTaken       = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound     = errors.New("user not found")
)

// Register creates a new user account.
func (s *Service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	if err := validateRegister(req); err != nil {
		return nil, err
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Check for duplicate email
	_, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err == nil {
		return nil, ErrEmailTaken
	}

	hash, err := password.Hash(req.Password)
	if err != nil {
		s.logger.Error("failed to hash password", "error", err)
		return nil, fmt.Errorf("auth: register: %w", err)
	}

	user, err := s.repo.CreateUser(ctx, sqlc.CreateUserParams{
		FullName:     strings.TrimSpace(req.FullName),
		Email:        req.Email,
		PasswordHash: hash,
		Role:         sqlc.UserRolePharmacist,
	})
	if err != nil {
		s.logger.Error("failed to create user", "error", err)
		return nil, fmt.Errorf("auth: create user: %w", err)
	}

	token, err := s.jwtManager.Generate(user.ID, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("auth: generate token: %w", err)
	}

	return &AuthResponse{
		Token: token,
		User: UserInfo{
			ID:       user.ID,
			FullName: user.FullName,
			Email:    user.Email,
			Role:     string(user.Role),
		},
	}, nil
}

// Login authenticates a user and returns a JWT.
func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	if err := validateLogin(req); err != nil {
		return nil, err
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := password.Compare(user.PasswordHash, req.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.jwtManager.Generate(user.ID, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("auth: generate token: %w", err)
	}

	return &AuthResponse{
		Token: token,
		User: UserInfo{
			ID:       user.ID,
			FullName: user.FullName,
			Email:    user.Email,
			Role:     string(user.Role),
		},
	}, nil
}

// GetCurrentUser returns the user info for the given user ID.
func (s *Service) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*UserInfo, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return &UserInfo{
		ID:       user.ID,
		FullName: user.FullName,
		Email:    user.Email,
		Role:     string(user.Role),
	}, nil
}

func validateRegister(req RegisterRequest) error {
	if strings.TrimSpace(req.FullName) == "" {
		return fmt.Errorf("full_name is required")
	}
	if utf8.RuneCountInString(req.FullName) > 200 {
		return fmt.Errorf("full_name must be at most 200 characters")
	}
	if strings.TrimSpace(req.Email) == "" {
		return fmt.Errorf("email is required")
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return fmt.Errorf("invalid email format")
	}
	if len(req.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	if len(req.Password) > 128 {
		return fmt.Errorf("password must be at most 128 characters")
	}
	return nil
}

func validateLogin(req LoginRequest) error {
	if strings.TrimSpace(req.Email) == "" {
		return fmt.Errorf("email is required")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}
