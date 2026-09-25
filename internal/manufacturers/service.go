package manufacturers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/drug-verification/server/internal/database/sqlc"
)

// Repository defines the data access interface for manufacturers.
type Repository interface {
	CreateManufacturer(ctx context.Context, params sqlc.CreateManufacturerParams) (sqlc.Manufacturer, error)
	GetManufacturer(ctx context.Context, id uuid.UUID) (sqlc.Manufacturer, error)
	ListManufacturers(ctx context.Context, params sqlc.ListManufacturersParams) ([]sqlc.Manufacturer, error)
	CountManufacturers(ctx context.Context) (int64, error)
	UpdateManufacturer(ctx context.Context, params sqlc.UpdateManufacturerParams) (sqlc.Manufacturer, error)
	DeleteManufacturer(ctx context.Context, id uuid.UUID) error
}

// Service implements manufacturer business logic.
type Service struct {
	repo   Repository
	logger *slog.Logger
}

// NewService creates a new manufacturer service.
func NewService(repo Repository, logger *slog.Logger) *Service {
	return &Service{repo: repo, logger: logger}
}

// CreateRequest represents a manufacturer creation request.
type CreateRequest struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

// UpdateRequest represents a manufacturer update request.
type UpdateRequest struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

// ManufacturerResponse is the API response for a manufacturer.
type ManufacturerResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

// Sentinel errors.
var (
	ErrManufacturerNotFound = errors.New("manufacturer not found")
	ErrManufacturerInUse    = errors.New("manufacturer is referenced by existing drugs")
)

func toResponse(m sqlc.Manufacturer) ManufacturerResponse {
	return ManufacturerResponse{
		ID:        m.ID,
		Name:      m.Name,
		Address:   m.Address,
		CreatedAt: m.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: m.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// Create adds a new manufacturer.
func (s *Service) Create(ctx context.Context, req CreateRequest) (*ManufacturerResponse, error) {
	if err := validateCreate(req); err != nil {
		return nil, err
	}

	m, err := s.repo.CreateManufacturer(ctx, sqlc.CreateManufacturerParams{
		Name:    strings.TrimSpace(req.Name),
		Address: strings.TrimSpace(req.Address),
	})
	if err != nil {
		s.logger.Error("failed to create manufacturer", "error", err)
		return nil, fmt.Errorf("manufacturer: create: %w", err)
	}

	resp := toResponse(m)
	return &resp, nil
}

// Get retrieves a manufacturer by ID.
func (s *Service) Get(ctx context.Context, id uuid.UUID) (*ManufacturerResponse, error) {
	m, err := s.repo.GetManufacturer(ctx, id)
	if err != nil {
		return nil, ErrManufacturerNotFound
	}

	resp := toResponse(m)
	return &resp, nil
}

// ListParams holds list/pagination parameters.
type ListParams struct {
	Page  int
	Limit int
}

// List retrieves manufacturers with pagination.
func (s *Service) List(ctx context.Context, params ListParams) ([]ManufacturerResponse, int64, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 || params.Limit > 100 {
		params.Limit = 20
	}
	offset := (params.Page - 1) * params.Limit

	manufacturers, err := s.repo.ListManufacturers(ctx, sqlc.ListManufacturersParams{
		Limit:  int32(params.Limit),
		Offset: int32(offset),
	})
	if err != nil {
		s.logger.Error("failed to list manufacturers", "error", err)
		return nil, 0, fmt.Errorf("manufacturer: list: %w", err)
	}

	total, err := s.repo.CountManufacturers(ctx)
	if err != nil {
		s.logger.Error("failed to count manufacturers", "error", err)
		return nil, 0, fmt.Errorf("manufacturer: count: %w", err)
	}

	result := make([]ManufacturerResponse, len(manufacturers))
	for i, m := range manufacturers {
		result[i] = toResponse(m)
	}

	return result, total, nil
}

// Update modifies an existing manufacturer.
func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*ManufacturerResponse, error) {
	if err := validateCreate(req.toCreateRequest()); err != nil {
		return nil, err
	}

	m, err := s.repo.UpdateManufacturer(ctx, sqlc.UpdateManufacturerParams{
		ID:      id,
		Name:    strings.TrimSpace(req.Name),
		Address: strings.TrimSpace(req.Address),
	})
	if err != nil {
		return nil, ErrManufacturerNotFound
	}

	resp := toResponse(m)
	return &resp, nil
}

// Delete removes a manufacturer.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	// Verify it exists
	_, err := s.repo.GetManufacturer(ctx, id)
	if err != nil {
		return ErrManufacturerNotFound
	}

	if err := s.repo.DeleteManufacturer(ctx, id); err != nil {
		// Foreign key violation means manufacturer is in use
		if strings.Contains(err.Error(), "violates foreign key") {
			return ErrManufacturerInUse
		}
		s.logger.Error("failed to delete manufacturer", "error", err)
		return fmt.Errorf("manufacturer: delete: %w", err)
	}

	return nil
}

func (r UpdateRequest) toCreateRequest() CreateRequest {
	return CreateRequest{Name: r.Name, Address: r.Address}
}

func validateCreate(req CreateRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if utf8.RuneCountInString(req.Name) > 300 {
		return fmt.Errorf("name must be at most 300 characters")
	}
	if utf8.RuneCountInString(req.Address) > 500 {
		return fmt.Errorf("address must be at most 500 characters")
	}
	return nil
}
