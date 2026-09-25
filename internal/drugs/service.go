package drugs

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/drug-verification/server/internal/database/sqlc"
)

// Repository defines the data access interface for drugs.
type Repository interface {
	CreateDrug(ctx context.Context, params sqlc.CreateDrugParams) (sqlc.Drug, error)
	GetDrug(ctx context.Context, id uuid.UUID) (sqlc.GetDrugRow, error)
	ListDrugs(ctx context.Context, params sqlc.ListDrugsParams) ([]sqlc.ListDrugsRow, error)
	CountDrugs(ctx context.Context, params sqlc.CountDrugsParams) (int64, error)
	UpdateDrug(ctx context.Context, params sqlc.UpdateDrugParams) (sqlc.Drug, error)
	UpdateDrugStatus(ctx context.Context, params sqlc.UpdateDrugStatusParams) (sqlc.Drug, error)
	GetDrugByManufacturerAndBatch(ctx context.Context, params sqlc.GetDrugByManufacturerAndBatchParams) (uuid.UUID, error)
	GetExpiredActiveDrugs(ctx context.Context) ([]sqlc.GetExpiredActiveDrugsRow, error)

	// Drug units
	CreateDrugUnit(ctx context.Context, params sqlc.CreateDrugUnitParams) (sqlc.DrugUnit, error)
	ListDrugUnitsByDrug(ctx context.Context, params sqlc.ListDrugUnitsByDrugParams) ([]sqlc.DrugUnit, error)
	CountDrugUnitsByDrug(ctx context.Context, drugID uuid.UUID) (int64, error)
	GetDrugUnit(ctx context.Context, id uuid.UUID) (sqlc.DrugUnit, error)

	// Transaction support
	BeginTx(ctx context.Context) (pgx.Tx, error)
	WithTx(tx pgx.Tx) *sqlc.Queries
}

// Service implements drug business logic.
type Service struct {
	repo   Repository
	logger *slog.Logger
}

// NewService creates a new drug service.
func NewService(repo Repository, logger *slog.Logger) *Service {
	return &Service{repo: repo, logger: logger}
}

// CreateRequest represents a drug creation request.
type CreateRequest struct {
	Name              string    `json:"name"`
	GenericName       string    `json:"generic_name"`
	ManufacturerID    string    `json:"manufacturer_id"`
	BatchNumber       string    `json:"batch_number"`
	NafdacNumber      string    `json:"nafdac_number"`
	ManufacturingDate string    `json:"manufacturing_date"`
	ExpiryDate        string    `json:"expiry_date"`
	Quantity          int       `json:"quantity"`
}

// UpdateRequest represents a drug update request.
type UpdateRequest struct {
	Name              string `json:"name"`
	GenericName       string `json:"generic_name"`
	BatchNumber       string `json:"batch_number"`
	NafdacNumber      string `json:"nafdac_number"`
	ManufacturingDate string `json:"manufacturing_date"`
	ExpiryDate        string `json:"expiry_date"`
	Quantity          int    `json:"quantity"`
	Status            string `json:"status"`
}

// GenerateUnitsRequest represents a drug unit generation request.
type GenerateUnitsRequest struct {
	Quantity int `json:"quantity"`
}

// DrugResponse is the API response for a drug.
type DrugResponse struct {
	ID                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	GenericName       string    `json:"generic_name"`
	ManufacturerID    uuid.UUID `json:"manufacturer_id"`
	ManufacturerName  string    `json:"manufacturer_name"`
	BatchNumber       string    `json:"batch_number"`
	NafdacNumber      string    `json:"nafdac_number"`
	ManufacturingDate string    `json:"manufacturing_date"`
	ExpiryDate        string    `json:"expiry_date"`
	Quantity          int       `json:"quantity"`
	Status            string    `json:"status"`
	CreatedBy         uuid.UUID `json:"created_by"`
	CreatedAt         string    `json:"created_at"`
	UpdatedAt         string    `json:"updated_at"`
}

// DrugUnitResponse is the API response for a drug unit.
type DrugUnitResponse struct {
	ID           uuid.UUID `json:"id"`
	DrugID       uuid.UUID `json:"drug_id"`
	SerialNumber string    `json:"serial_number"`
	Status       string    `json:"status"`
	CreatedAt    string    `json:"created_at"`
}

// ListParams holds drug listing parameters.
type ListParams struct {
	Page           int
	Limit          int
	Name           string
	GenericName    string
	BatchNumber    string
	Status         string
	ManufacturerID string
	ExpiryBefore   string
	ExpiryAfter    string
}

// Sentinel errors.
var (
	ErrDrugNotFound     = errors.New("drug not found")
	ErrDuplicateBatch   = errors.New("duplicate batch number for this manufacturer")
	ErrDrugUnitNotFound = errors.New("drug unit not found")
)

// Create adds a new drug.
func (s *Service) Create(ctx context.Context, req CreateRequest, createdBy uuid.UUID) (*DrugResponse, error) {
	if err := validateCreate(req); err != nil {
		return nil, err
	}

	manufacturerID, err := uuid.Parse(req.ManufacturerID)
	if err != nil {
		return nil, fmt.Errorf("invalid manufacturer_id format")
	}

	mfgDate, _ := time.Parse("2006-01-02", req.ManufacturingDate)
	expDate, _ := time.Parse("2006-01-02", req.ExpiryDate)

	// Check duplicate batch for same manufacturer
	_, err = s.repo.GetDrugByManufacturerAndBatch(ctx, sqlc.GetDrugByManufacturerAndBatchParams{
		ManufacturerID: manufacturerID,
		BatchNumber:    strings.TrimSpace(req.BatchNumber),
	})
	if err == nil {
		return nil, ErrDuplicateBatch
	}

	drug, err := s.repo.CreateDrug(ctx, sqlc.CreateDrugParams{
		Name:              strings.TrimSpace(req.Name),
		GenericName:       strings.TrimSpace(req.GenericName),
		ManufacturerID:    manufacturerID,
		BatchNumber:       strings.TrimSpace(req.BatchNumber),
		NafdacNumber:      strings.TrimSpace(req.NafdacNumber),
		ManufacturingDate: mfgDate,
		ExpiryDate:        expDate,
		Quantity:          int32(req.Quantity),
		Status:            sqlc.DrugStatusActive,
		CreatedBy:         createdBy,
	})
	if err != nil {
		s.logger.Error("failed to create drug", "error", err)
		return nil, fmt.Errorf("drug: create: %w", err)
	}

	// Fetch with manufacturer name
	full, err := s.repo.GetDrug(ctx, drug.ID)
	if err != nil {
		return nil, fmt.Errorf("drug: fetch created: %w", err)
	}

	resp := drugRowToResponse(full)
	return &resp, nil
}

// Get retrieves a drug by ID.
func (s *Service) Get(ctx context.Context, id uuid.UUID) (*DrugResponse, error) {
	drug, err := s.repo.GetDrug(ctx, id)
	if err != nil {
		return nil, ErrDrugNotFound
	}

	resp := drugRowToResponse(drug)
	return &resp, nil
}

// List retrieves drugs with pagination and filtering.
func (s *Service) List(ctx context.Context, params ListParams) ([]DrugResponse, int64, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 || params.Limit > 100 {
		params.Limit = 20
	}
	offset := (params.Page - 1) * params.Limit

	listParams := sqlc.ListDrugsParams{
		Limit:  int32(params.Limit),
		Offset: int32(offset),
	}
	countParams := sqlc.CountDrugsParams{}

	// Apply filters
	if params.Name != "" {
		listParams.Name = pgtype.Text{String: params.Name, Valid: true}
		countParams.Name = pgtype.Text{String: params.Name, Valid: true}
	}
	if params.GenericName != "" {
		listParams.GenericName = pgtype.Text{String: params.GenericName, Valid: true}
		countParams.GenericName = pgtype.Text{String: params.GenericName, Valid: true}
	}
	if params.BatchNumber != "" {
		listParams.BatchNumber = pgtype.Text{String: params.BatchNumber, Valid: true}
		countParams.BatchNumber = pgtype.Text{String: params.BatchNumber, Valid: true}
	}
	if params.Status != "" {
		st := sqlc.NullDrugStatus{DrugStatus: sqlc.DrugStatus(params.Status), Valid: true}
		listParams.Status = st
		countParams.Status = st
	}
	if params.ManufacturerID != "" {
		if mid, err := uuid.Parse(params.ManufacturerID); err == nil {
			u := pgtype.UUID{Bytes: mid, Valid: true}
			listParams.ManufacturerID = u
			countParams.ManufacturerID = u
		}
	}
	if params.ExpiryBefore != "" {
		if t, err := time.Parse("2006-01-02", params.ExpiryBefore); err == nil {
			d := pgtype.Date{Time: t, Valid: true}
			listParams.ExpiryBefore = d
			countParams.ExpiryBefore = d
		}
	}
	if params.ExpiryAfter != "" {
		if t, err := time.Parse("2006-01-02", params.ExpiryAfter); err == nil {
			d := pgtype.Date{Time: t, Valid: true}
			listParams.ExpiryAfter = d
			countParams.ExpiryAfter = d
		}
	}

	drugs, err := s.repo.ListDrugs(ctx, listParams)
	if err != nil {
		s.logger.Error("failed to list drugs", "error", err)
		return nil, 0, fmt.Errorf("drug: list: %w", err)
	}

	total, err := s.repo.CountDrugs(ctx, countParams)
	if err != nil {
		s.logger.Error("failed to count drugs", "error", err)
		return nil, 0, fmt.Errorf("drug: count: %w", err)
	}

	result := make([]DrugResponse, len(drugs))
	for i, d := range drugs {
		result[i] = listRowToResponse(d)
	}

	return result, total, nil
}

// Update modifies an existing drug.
func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateRequest) (*DrugResponse, error) {
	if err := validateUpdate(req); err != nil {
		return nil, err
	}

	// Verify drug exists
	existing, err := s.repo.GetDrug(ctx, id)
	if err != nil {
		return nil, ErrDrugNotFound
	}

	mfgDate, _ := time.Parse("2006-01-02", req.ManufacturingDate)
	expDate, _ := time.Parse("2006-01-02", req.ExpiryDate)

	status := sqlc.DrugStatus(req.Status)
	if req.Status == "" {
		status = existing.Status
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = existing.Name
	}

	_, err = s.repo.UpdateDrug(ctx, sqlc.UpdateDrugParams{
		ID:                id,
		Name:              name,
		GenericName:       coalesce(strings.TrimSpace(req.GenericName), existing.GenericName),
		BatchNumber:       coalesce(strings.TrimSpace(req.BatchNumber), existing.BatchNumber),
		NafdacNumber:      coalesce(strings.TrimSpace(req.NafdacNumber), existing.NafdacNumber),
		ManufacturingDate: coalesceTime(mfgDate, existing.ManufacturingDate),
		ExpiryDate:        coalesceTime(expDate, existing.ExpiryDate),
		Quantity:          coalesceInt32(int32(req.Quantity), existing.Quantity),
		Status:            status,
	})
	if err != nil {
		s.logger.Error("failed to update drug", "error", err)
		return nil, fmt.Errorf("drug: update: %w", err)
	}

	updated, err := s.repo.GetDrug(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("drug: fetch updated: %w", err)
	}

	resp := drugRowToResponse(updated)
	return &resp, nil
}

// UpdateStatus changes a drug's status.
func (s *Service) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	if !isValidDrugStatus(status) {
		return fmt.Errorf("invalid status: %s", status)
	}

	_, err := s.repo.UpdateDrugStatus(ctx, sqlc.UpdateDrugStatusParams{
		ID:     id,
		Status: sqlc.DrugStatus(status),
	})
	if err != nil {
		return ErrDrugNotFound
	}
	return nil
}

// GenerateUnits creates drug units with unique serial numbers.
func (s *Service) GenerateUnits(ctx context.Context, drugID uuid.UUID, quantity int) ([]DrugUnitResponse, error) {
	if quantity < 1 {
		return nil, fmt.Errorf("quantity must be at least 1")
	}
	if quantity > 1000 {
		return nil, fmt.Errorf("quantity must be at most 1000 per request")
	}

	// Verify drug exists
	_, err := s.repo.GetDrug(ctx, drugID)
	if err != nil {
		return nil, ErrDrugNotFound
	}

	// Use transaction for atomic unit creation
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("drug: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	queries := s.repo.WithTx(tx)

	units := make([]DrugUnitResponse, 0, quantity)
	for i := 0; i < quantity; i++ {
		serial, err := generateSerialNumber()
		if err != nil {
			return nil, fmt.Errorf("drug: generate serial: %w", err)
		}

		unit, err := queries.CreateDrugUnit(ctx, sqlc.CreateDrugUnitParams{
			DrugID:       drugID,
			SerialNumber: serial,
			Status:       sqlc.DrugUnitStatusActive,
		})
		if err != nil {
			return nil, fmt.Errorf("drug: create unit: %w", err)
		}

		units = append(units, DrugUnitResponse{
			ID:           unit.ID,
			DrugID:       unit.DrugID,
			SerialNumber: unit.SerialNumber,
			Status:       string(unit.Status),
			CreatedAt:    unit.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("drug: commit tx: %w", err)
	}

	return units, nil
}

// ListUnits retrieves drug units for a specific drug.
func (s *Service) ListUnits(ctx context.Context, drugID uuid.UUID, page, limit int) ([]DrugUnitResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	units, err := s.repo.ListDrugUnitsByDrug(ctx, sqlc.ListDrugUnitsByDrugParams{
		DrugID: drugID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("drug: list units: %w", err)
	}

	total, err := s.repo.CountDrugUnitsByDrug(ctx, drugID)
	if err != nil {
		return nil, 0, fmt.Errorf("drug: count units: %w", err)
	}

	result := make([]DrugUnitResponse, len(units))
	for i, u := range units {
		result[i] = DrugUnitResponse{
			ID:           u.ID,
			DrugID:       u.DrugID,
			SerialNumber: u.SerialNumber,
			Status:       string(u.Status),
			CreatedAt:    u.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	return result, total, nil
}

// GetUnit retrieves a single drug unit.
func (s *Service) GetUnit(ctx context.Context, id uuid.UUID) (*DrugUnitResponse, error) {
	u, err := s.repo.GetDrugUnit(ctx, id)
	if err != nil {
		return nil, ErrDrugUnitNotFound
	}

	return &DrugUnitResponse{
		ID:           u.ID,
		DrugID:       u.DrugID,
		SerialNumber: u.SerialNumber,
		Status:       string(u.Status),
		CreatedAt:    u.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}

// generateSerialNumber creates a cryptographically random serial number.
// Format: 8 uppercase hex characters (e.g., "A8F72K91").
func generateSerialNumber() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return strings.ToUpper(hex.EncodeToString(b)), nil
}

func drugRowToResponse(d sqlc.GetDrugRow) DrugResponse {
	return DrugResponse{
		ID:                d.ID,
		Name:              d.Name,
		GenericName:       d.GenericName,
		ManufacturerID:    d.ManufacturerID,
		ManufacturerName:  d.ManufacturerName,
		BatchNumber:       d.BatchNumber,
		NafdacNumber:      d.NafdacNumber,
		ManufacturingDate: d.ManufacturingDate.Format("2006-01-02"),
		ExpiryDate:        d.ExpiryDate.Format("2006-01-02"),
		Quantity:          int(d.Quantity),
		Status:            string(d.Status),
		CreatedBy:         d.CreatedBy,
		CreatedAt:         d.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:         d.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func listRowToResponse(d sqlc.ListDrugsRow) DrugResponse {
	return DrugResponse{
		ID:                d.ID,
		Name:              d.Name,
		GenericName:       d.GenericName,
		ManufacturerID:    d.ManufacturerID,
		ManufacturerName:  d.ManufacturerName,
		BatchNumber:       d.BatchNumber,
		NafdacNumber:      d.NafdacNumber,
		ManufacturingDate: d.ManufacturingDate.Format("2006-01-02"),
		ExpiryDate:        d.ExpiryDate.Format("2006-01-02"),
		Quantity:          int(d.Quantity),
		Status:            string(d.Status),
		CreatedBy:         d.CreatedBy,
		CreatedAt:         d.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:         d.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func validateCreate(req CreateRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if utf8.RuneCountInString(req.Name) > 300 {
		return fmt.Errorf("name must be at most 300 characters")
	}
	if strings.TrimSpace(req.ManufacturerID) == "" {
		return fmt.Errorf("manufacturer_id is required")
	}
	if _, err := uuid.Parse(req.ManufacturerID); err != nil {
		return fmt.Errorf("invalid manufacturer_id format")
	}
	if strings.TrimSpace(req.BatchNumber) == "" {
		return fmt.Errorf("batch_number is required")
	}
	if strings.TrimSpace(req.ManufacturingDate) == "" {
		return fmt.Errorf("manufacturing_date is required")
	}
	mfgDate, err := time.Parse("2006-01-02", req.ManufacturingDate)
	if err != nil {
		return fmt.Errorf("manufacturing_date must be in YYYY-MM-DD format")
	}
	if strings.TrimSpace(req.ExpiryDate) == "" {
		return fmt.Errorf("expiry_date is required")
	}
	expDate, err := time.Parse("2006-01-02", req.ExpiryDate)
	if err != nil {
		return fmt.Errorf("expiry_date must be in YYYY-MM-DD format")
	}
	if mfgDate.After(expDate) {
		return fmt.Errorf("manufacturing_date cannot be after expiry_date")
	}
	if req.Quantity < 0 {
		return fmt.Errorf("quantity cannot be negative")
	}
	return nil
}

func validateUpdate(req UpdateRequest) error {
	if req.ManufacturingDate != "" {
		if _, err := time.Parse("2006-01-02", req.ManufacturingDate); err != nil {
			return fmt.Errorf("manufacturing_date must be in YYYY-MM-DD format")
		}
	}
	if req.ExpiryDate != "" {
		if _, err := time.Parse("2006-01-02", req.ExpiryDate); err != nil {
			return fmt.Errorf("expiry_date must be in YYYY-MM-DD format")
		}
	}
	if req.ManufacturingDate != "" && req.ExpiryDate != "" {
		mfg, _ := time.Parse("2006-01-02", req.ManufacturingDate)
		exp, _ := time.Parse("2006-01-02", req.ExpiryDate)
		if mfg.After(exp) {
			return fmt.Errorf("manufacturing_date cannot be after expiry_date")
		}
	}
	if req.Quantity < 0 {
		return fmt.Errorf("quantity cannot be negative")
	}
	if req.Status != "" && !isValidDrugStatus(req.Status) {
		return fmt.Errorf("invalid status: must be one of active, expired, recalled, suspended")
	}
	return nil
}

func isValidDrugStatus(s string) bool {
	switch sqlc.DrugStatus(s) {
	case sqlc.DrugStatusActive, sqlc.DrugStatusExpired,
		sqlc.DrugStatusRecalled, sqlc.DrugStatusSuspended:
		return true
	}
	return false
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func coalesce(val, fallback string) string {
	if val != "" {
		return val
	}
	return fallback
}

func coalesceTime(val, fallback time.Time) time.Time {
	if val.IsZero() {
		return fallback
	}
	return val
}

func coalesceInt32(val, fallback int32) int32 {
	if val == 0 {
		return fallback
	}
	return val
}
