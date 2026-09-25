package verification_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/drug-verification/server/internal/database/sqlc"
	"github.com/drug-verification/server/internal/verification"
)

type mockRepository struct {
	unitRow     sqlc.GetDrugUnitBySerialRow
	getErr      error
	loggedCalls []sqlc.CreateVerificationLogParams
}

func (m *mockRepository) GetDrugUnitBySerial(ctx context.Context, serial string) (sqlc.GetDrugUnitBySerialRow, error) {
	if m.getErr != nil {
		return sqlc.GetDrugUnitBySerialRow{}, m.getErr
	}
	return m.unitRow, nil
}

func (m *mockRepository) CreateVerificationLog(ctx context.Context, params sqlc.CreateVerificationLogParams) (sqlc.VerificationLog, error) {
	m.loggedCalls = append(m.loggedCalls, params)
	return sqlc.VerificationLog{
		ID:           uuid.New(),
		SerialNumber: params.SerialNumber,
		Verified:     params.Verified,
		Result:       params.Result,
		VerifiedAt:   time.Now(),
	}, nil
}

func (m *mockRepository) ListVerificationLogs(ctx context.Context, params sqlc.ListVerificationLogsParams) ([]sqlc.ListVerificationLogsRow, error) {
	return nil, nil
}

func (m *mockRepository) CountVerificationLogs(ctx context.Context, params sqlc.CountVerificationLogsParams) (int64, error) {
	return 0, nil
}

func (m *mockRepository) ListVerificationLogsByDrug(ctx context.Context, params sqlc.ListVerificationLogsByDrugParams) ([]sqlc.ListVerificationLogsByDrugRow, error) {
	return nil, nil
}

func (m *mockRepository) CountVerificationLogsByDrug(ctx context.Context, params sqlc.CountVerificationLogsByDrugParams) (int64, error) {
	return 0, nil
}

func newDiscardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestVerify_ValidDrugUnit(t *testing.T) {
	unitID := uuid.New()
	mockRepo := &mockRepository{
		unitRow: sqlc.GetDrugUnitBySerialRow{
			ID:                    unitID,
			DrugID:                uuid.New(),
			SerialNumber:          "AMOX-2026-ABCD",
			Status:                sqlc.DrugUnitStatusActive,
			DrugName:              "Amoxicillin 500mg",
			DrugGenericName:       "Amoxicillin",
			DrugBatchNumber:       "BATCH-100",
			DrugNafdacNumber:      "A4-1234",
			DrugManufacturingDate: time.Now().Add(-30 * 24 * time.Hour),
			DrugExpiryDate:        time.Now().Add(365 * 24 * time.Hour),
			DrugStatus:            sqlc.DrugStatusActive,
			ManufacturerName:      "PharmaCorp Ltd",
		},
	}

	service := verification.NewService(mockRepo, newDiscardLogger())
	res, err := service.Verify(context.Background(), verification.VerifyRequest{
		SerialNumber: "AMOX-2026-ABCD",
		IPAddress:    "127.0.0.1",
		UserAgent:    "Go-Test",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !res.Verified {
		t.Errorf("expected Verified == true, got false")
	}
	if res.Result != "verified" {
		t.Errorf("expected Result == 'verified', got %q", res.Result)
	}
	if res.Drug == nil || res.Drug.Name != "Amoxicillin 500mg" {
		t.Errorf("expected drug info to be populated with Amoxicillin 500mg")
	}

	if len(mockRepo.loggedCalls) != 1 {
		t.Fatalf("expected 1 audit log created, got %d", len(mockRepo.loggedCalls))
	}
	if !mockRepo.loggedCalls[0].Verified {
		t.Errorf("expected audit log verified to be true")
	}
}

func TestVerify_NotFound(t *testing.T) {
	mockRepo := &mockRepository{
		getErr: errors.New("no rows in result set"),
	}

	service := verification.NewService(mockRepo, newDiscardLogger())
	res, err := service.Verify(context.Background(), verification.VerifyRequest{
		SerialNumber: "UNKNOWN-SERIAL-999",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Verified {
		t.Errorf("expected Verified == false for non-existent serial")
	}
	if res.Result != "not_found" {
		t.Errorf("expected Result == 'not_found', got %q", res.Result)
	}

	if len(mockRepo.loggedCalls) != 1 {
		t.Fatalf("expected 1 audit log created, got %d", len(mockRepo.loggedCalls))
	}
	if mockRepo.loggedCalls[0].Result != sqlc.VerificationResultNotFound {
		t.Errorf("expected audit log result to be 'not_found', got %v", mockRepo.loggedCalls[0].Result)
	}
}

func TestVerify_ExpiredDrug(t *testing.T) {
	unitID := uuid.New()
	mockRepo := &mockRepository{
		unitRow: sqlc.GetDrugUnitBySerialRow{
			ID:                    unitID,
			DrugID:                uuid.New(),
			SerialNumber:          "AMOX-EXPIRED",
			Status:                sqlc.DrugUnitStatusActive,
			DrugName:              "Amoxicillin 500mg",
			DrugGenericName:       "Amoxicillin",
			DrugBatchNumber:       "BATCH-EXPIRED",
			DrugNafdacNumber:      "A4-1234",
			DrugManufacturingDate: time.Now().Add(-400 * 24 * time.Hour),
			DrugExpiryDate:        time.Now().Add(-10 * 24 * time.Hour), // Expired 10 days ago
			DrugStatus:            sqlc.DrugStatusActive,
			ManufacturerName:      "PharmaCorp Ltd",
		},
	}

	service := verification.NewService(mockRepo, newDiscardLogger())
	res, err := service.Verify(context.Background(), verification.VerifyRequest{
		SerialNumber: "AMOX-EXPIRED",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Verified {
		t.Errorf("expected Verified == false for expired drug")
	}
	if res.Result != "expired" {
		t.Errorf("expected Result == 'expired', got %q", res.Result)
	}
}

func TestVerify_RecalledBatch(t *testing.T) {
	unitID := uuid.New()
	mockRepo := &mockRepository{
		unitRow: sqlc.GetDrugUnitBySerialRow{
			ID:                    unitID,
			DrugID:                uuid.New(),
			SerialNumber:          "AMOX-RECALLED",
			Status:                sqlc.DrugUnitStatusActive,
			DrugName:              "Amoxicillin 500mg",
			DrugGenericName:       "Amoxicillin",
			DrugBatchNumber:       "BATCH-RECALLED",
			DrugNafdacNumber:      "A4-1234",
			DrugManufacturingDate: time.Now().Add(-10 * 24 * time.Hour),
			DrugExpiryDate:        time.Now().Add(200 * 24 * time.Hour),
			DrugStatus:            sqlc.DrugStatusRecalled,
			ManufacturerName:      "PharmaCorp Ltd",
		},
	}

	service := verification.NewService(mockRepo, newDiscardLogger())
	res, err := service.Verify(context.Background(), verification.VerifyRequest{
		SerialNumber: "AMOX-RECALLED",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Verified {
		t.Errorf("expected Verified == false for recalled drug")
	}
	if res.Result != "recalled" {
		t.Errorf("expected Result == 'recalled', got %q", res.Result)
	}
}
