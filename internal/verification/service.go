package verification

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/drug-verification/server/internal/database/sqlc"
)

// Repository defines the data access interface for verification.
type Repository interface {
	GetDrugUnitBySerial(ctx context.Context, serial string) (sqlc.GetDrugUnitBySerialRow, error)
	CreateVerificationLog(ctx context.Context, params sqlc.CreateVerificationLogParams) (sqlc.VerificationLog, error)
	ListVerificationLogs(ctx context.Context, params sqlc.ListVerificationLogsParams) ([]sqlc.ListVerificationLogsRow, error)
	CountVerificationLogs(ctx context.Context, params sqlc.CountVerificationLogsParams) (int64, error)
	ListVerificationLogsByDrug(ctx context.Context, params sqlc.ListVerificationLogsByDrugParams) ([]sqlc.ListVerificationLogsByDrugRow, error)
	CountVerificationLogsByDrug(ctx context.Context, params sqlc.CountVerificationLogsByDrugParams) (int64, error)
}

// Service implements verification business logic.
type Service struct {
	repo   Repository
	logger *slog.Logger
}

// NewService creates a new verification service.
func NewService(repo Repository, logger *slog.Logger) *Service {
	return &Service{repo: repo, logger: logger}
}

// VerifyResult represents the outcome of a drug verification.
type VerifyResult struct {
	Verified bool           `json:"verified"`
	Result   string         `json:"result"`
	Message  string         `json:"message,omitempty"`
	Drug     *DrugInfo      `json:"drug,omitempty"`
	DrugUnit *DrugUnitInfo  `json:"drug_unit,omitempty"`
}

// DrugInfo holds drug information returned during verification.
type DrugInfo struct {
	Name              string `json:"name"`
	GenericName       string `json:"generic_name"`
	Manufacturer      string `json:"manufacturer"`
	BatchNumber       string `json:"batch_number"`
	NafdacNumber      string `json:"nafdac_number"`
	ManufacturingDate string `json:"manufacturing_date"`
	ExpiryDate        string `json:"expiry_date"`
}

// DrugUnitInfo holds drug unit information returned during verification.
type DrugUnitInfo struct {
	SerialNumber string `json:"serial_number"`
	Status       string `json:"status"`
}

// VerificationLogResponse is the API response for a verification log entry.
type VerificationLogResponse struct {
	ID           uuid.UUID  `json:"id"`
	DrugUnitID   *uuid.UUID `json:"drug_unit_id,omitempty"`
	SerialNumber string     `json:"serial_number"`
	Verified     bool       `json:"verified"`
	Result       string     `json:"result"`
	VerifiedAt   string     `json:"verified_at"`
}

// LogListParams holds verification log listing parameters.
type LogListParams struct {
	Page         int
	Limit        int
	SerialNumber string
	Result       string
	FromDate     string
	ToDate       string
}

// VerifyRequest holds the context for a verification attempt.
type VerifyRequest struct {
	SerialNumber string
	IPAddress    string
	UserAgent    string
}

// Verify performs drug verification by serial number.
// This is the core business logic of the system.
func (s *Service) Verify(ctx context.Context, req VerifyRequest) (*VerifyResult, error) {
	serial := req.SerialNumber

	// Validate serial format
	if serial == "" || len(serial) > 100 {
		result := &VerifyResult{
			Verified: false,
			Result:   "invalid",
			Message:  "Invalid serial number format.",
		}
		s.logVerification(ctx, nil, serial, false, "invalid", req.IPAddress, req.UserAgent)
		return result, nil
	}

	// Look up drug unit with associated drug and manufacturer
	unit, err := s.repo.GetDrugUnitBySerial(ctx, serial)
	if err != nil {
		// Drug unit not found
		result := &VerifyResult{
			Verified: false,
			Result:   "not_found",
			Message:  "The drug could not be verified. No record found for this serial number.",
		}
		s.logVerification(ctx, nil, serial, false, "not_found", req.IPAddress, req.UserAgent)
		return result, nil
	}

	unitID := unit.ID

	// Check drug unit status
	switch unit.Status {
	case sqlc.DrugUnitStatusRecalled:
		result := &VerifyResult{
			Verified: false,
			Result:   "recalled",
			Message:  "This drug unit has been recalled. Do not use this product.",
		}
		s.logVerification(ctx, &unitID, serial, false, "recalled", req.IPAddress, req.UserAgent)
		return result, nil

	case sqlc.DrugUnitStatusSuspended:
		result := &VerifyResult{
			Verified: false,
			Result:   "suspended",
			Message:  "This drug unit has been suspended pending investigation.",
		}
		s.logVerification(ctx, &unitID, serial, false, "suspended", req.IPAddress, req.UserAgent)
		return result, nil

	case sqlc.DrugUnitStatusUsed:
		result := &VerifyResult{
			Verified: false,
			Result:   "invalid",
			Message:  "This drug unit has already been marked as used.",
		}
		s.logVerification(ctx, &unitID, serial, false, "invalid", req.IPAddress, req.UserAgent)
		return result, nil
	}

	// Check drug status
	switch unit.DrugStatus {
	case sqlc.DrugStatusRecalled:
		result := &VerifyResult{
			Verified: false,
			Result:   "recalled",
			Message:  "This drug batch has been recalled. Do not use this product.",
		}
		s.logVerification(ctx, &unitID, serial, false, "recalled", req.IPAddress, req.UserAgent)
		return result, nil

	case sqlc.DrugStatusSuspended:
		result := &VerifyResult{
			Verified: false,
			Result:   "suspended",
			Message:  "This drug batch has been suspended pending investigation.",
		}
		s.logVerification(ctx, &unitID, serial, false, "suspended", req.IPAddress, req.UserAgent)
		return result, nil
	}

	// Check expiry independently — NEVER rely solely on database status
	if unit.DrugExpiryDate.Before(time.Now().Truncate(24 * time.Hour)) {
		result := &VerifyResult{
			Verified: false,
			Result:   "expired",
			Message:  "This drug has expired. Do not use this product.",
			Drug: &DrugInfo{
				Name:              unit.DrugName,
				GenericName:       unit.DrugGenericName,
				Manufacturer:      unit.ManufacturerName,
				BatchNumber:       unit.DrugBatchNumber,
				NafdacNumber:      unit.DrugNafdacNumber,
				ManufacturingDate: unit.DrugManufacturingDate.Format("2006-01-02"),
				ExpiryDate:        unit.DrugExpiryDate.Format("2006-01-02"),
			},
			DrugUnit: &DrugUnitInfo{
				SerialNumber: unit.SerialNumber,
				Status:       string(unit.Status),
			},
		}
		s.logVerification(ctx, &unitID, serial, false, "expired", req.IPAddress, req.UserAgent)
		return result, nil
	}

	// All checks passed — drug record verified
	result := &VerifyResult{
		Verified: true,
		Result:   "verified",
		Message:  "Drug record verified. This serial number corresponds to a registered and active drug.",
		Drug: &DrugInfo{
			Name:              unit.DrugName,
			GenericName:       unit.DrugGenericName,
			Manufacturer:      unit.ManufacturerName,
			BatchNumber:       unit.DrugBatchNumber,
			NafdacNumber:      unit.DrugNafdacNumber,
			ManufacturingDate: unit.DrugManufacturingDate.Format("2006-01-02"),
			ExpiryDate:        unit.DrugExpiryDate.Format("2006-01-02"),
		},
		DrugUnit: &DrugUnitInfo{
			SerialNumber: unit.SerialNumber,
			Status:       string(unit.Status),
		},
	}
	s.logVerification(ctx, &unitID, serial, true, "verified", req.IPAddress, req.UserAgent)
	return result, nil
}

// logVerification records a verification attempt. Errors are logged but not propagated.
func (s *Service) logVerification(ctx context.Context, unitID *uuid.UUID, serial string, verified bool, result, ip, ua string) {
	params := sqlc.CreateVerificationLogParams{
		SerialNumber: serial,
		Verified:     verified,
		Result:       sqlc.VerificationResult(result),
		DrugUnitID:   uuidToPgUUID(unitID),
		IpAddress:    stringToPgText(ip),
		UserAgent:    stringToPgText(ua),
	}

	_, err := s.repo.CreateVerificationLog(ctx, params)
	if err != nil {
		s.logger.Error("failed to log verification", "serial", serial, "error", err)
	}
}

// ListLogs retrieves verification logs with pagination and filtering.
func (s *Service) ListLogs(ctx context.Context, params LogListParams) ([]VerificationLogResponse, int64, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 || params.Limit > 100 {
		params.Limit = 20
	}
	offset := (params.Page - 1) * params.Limit

	listParams := sqlc.ListVerificationLogsParams{
		Limit:  int32(params.Limit),
		Offset: int32(offset),
	}
	countParams := sqlc.CountVerificationLogsParams{}

	if params.SerialNumber != "" {
		listParams.SerialNumber = stringToPgText(params.SerialNumber)
		countParams.SerialNumber = stringToPgText(params.SerialNumber)
	}
	if params.Result != "" {
		r := sqlc.NullVerificationResult{
			VerificationResult: sqlc.VerificationResult(params.Result),
			Valid:              true,
		}
		listParams.Result = r
		countParams.Result = r
	}
	if params.FromDate != "" {
		if t, err := time.Parse(time.RFC3339, params.FromDate); err == nil {
			ts := timeToPgTimestamptz(t)
			listParams.FromDate = ts
			countParams.FromDate = ts
		}
	}
	if params.ToDate != "" {
		if t, err := time.Parse(time.RFC3339, params.ToDate); err == nil {
			ts := timeToPgTimestamptz(t)
			listParams.ToDate = ts
			countParams.ToDate = ts
		}
	}

	logs, err := s.repo.ListVerificationLogs(ctx, listParams)
	if err != nil {
		s.logger.Error("failed to list verification logs", "error", err)
		return nil, 0, fmt.Errorf("verification: list logs: %w", err)
	}

	total, err := s.repo.CountVerificationLogs(ctx, countParams)
	if err != nil {
		s.logger.Error("failed to count verification logs", "error", err)
		return nil, 0, fmt.Errorf("verification: count logs: %w", err)
	}

	result := make([]VerificationLogResponse, len(logs))
	for i, l := range logs {
		result[i] = VerificationLogResponse{
			ID:           l.ID,
			DrugUnitID:   pgUUIDToPtr(l.DrugUnitID),
			SerialNumber: l.SerialNumber,
			Verified:     l.Verified,
			Result:       string(l.Result),
			VerifiedAt:   l.VerifiedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	return result, total, nil
}

// ListLogsByDrug retrieves verification logs for a specific drug.
func (s *Service) ListLogsByDrug(ctx context.Context, drugID uuid.UUID, params LogListParams) ([]VerificationLogResponse, int64, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.Limit < 1 || params.Limit > 100 {
		params.Limit = 20
	}
	offset := (params.Page - 1) * params.Limit

	listParams := sqlc.ListVerificationLogsByDrugParams{
		DrugID: drugID,
		Limit:  int32(params.Limit),
		Offset: int32(offset),
	}
	countParams := sqlc.CountVerificationLogsByDrugParams{
		DrugID: drugID,
	}

	if params.Result != "" {
		r := sqlc.NullVerificationResult{
			VerificationResult: sqlc.VerificationResult(params.Result),
			Valid:              true,
		}
		listParams.Result = r
		countParams.Result = r
	}
	if params.FromDate != "" {
		if t, err := time.Parse(time.RFC3339, params.FromDate); err == nil {
			ts := timeToPgTimestamptz(t)
			listParams.FromDate = ts
			countParams.FromDate = ts
		}
	}
	if params.ToDate != "" {
		if t, err := time.Parse(time.RFC3339, params.ToDate); err == nil {
			ts := timeToPgTimestamptz(t)
			listParams.ToDate = ts
			countParams.ToDate = ts
		}
	}

	logs, err := s.repo.ListVerificationLogsByDrug(ctx, listParams)
	if err != nil {
		s.logger.Error("failed to list verification logs by drug", "error", err)
		return nil, 0, fmt.Errorf("verification: list logs: %w", err)
	}

	total, err := s.repo.CountVerificationLogsByDrug(ctx, countParams)
	if err != nil {
		return nil, 0, fmt.Errorf("verification: count logs by drug: %w", err)
	}

	result := make([]VerificationLogResponse, len(logs))
	for i, l := range logs {
		result[i] = VerificationLogResponse{
			ID:           l.ID,
			DrugUnitID:   pgUUIDToPtr(l.DrugUnitID),
			SerialNumber: l.SerialNumber,
			Verified:     l.Verified,
			Result:       string(l.Result),
			VerifiedAt:   l.VerifiedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	return result, total, nil
}

func uuidToPgUUID(u *uuid.UUID) pgtype.UUID {
	if u == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *u, Valid: true}
}

func pgUUIDToPtr(u pgtype.UUID) *uuid.UUID {
	if !u.Valid {
		return nil
	}
	id := uuid.UUID(u.Bytes)
	return &id
}

func stringToPgText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

func timeToPgTimestamptz(t time.Time) pgtype.Timestamptz {
	if t.IsZero() {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: t, Valid: true}
}

// Sentinel errors
var (
	ErrVerificationFailed = errors.New("verification failed")
)
