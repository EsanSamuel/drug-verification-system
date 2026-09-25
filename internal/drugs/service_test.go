package drugs_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/drug-verification/server/internal/database/sqlc"
	"github.com/drug-verification/server/internal/drugs"
)

type mockDrugRepo struct{}

func (m *mockDrugRepo) CreateDrug(ctx context.Context, params sqlc.CreateDrugParams) (sqlc.Drug, error) {
	return sqlc.Drug{}, nil
}
func (m *mockDrugRepo) GetDrug(ctx context.Context, id uuid.UUID) (sqlc.GetDrugRow, error) {
	return sqlc.GetDrugRow{}, nil
}
func (m *mockDrugRepo) ListDrugs(ctx context.Context, params sqlc.ListDrugsParams) ([]sqlc.ListDrugsRow, error) {
	return nil, nil
}
func (m *mockDrugRepo) CountDrugs(ctx context.Context, params sqlc.CountDrugsParams) (int64, error) {
	return 0, nil
}
func (m *mockDrugRepo) UpdateDrug(ctx context.Context, params sqlc.UpdateDrugParams) (sqlc.Drug, error) {
	return sqlc.Drug{}, nil
}
func (m *mockDrugRepo) UpdateDrugStatus(ctx context.Context, params sqlc.UpdateDrugStatusParams) (sqlc.Drug, error) {
	return sqlc.Drug{}, nil
}
func (m *mockDrugRepo) GetDrugByManufacturerAndBatch(ctx context.Context, params sqlc.GetDrugByManufacturerAndBatchParams) (uuid.UUID, error) {
	return uuid.Nil, nil
}
func (m *mockDrugRepo) GetExpiredActiveDrugs(ctx context.Context) ([]sqlc.GetExpiredActiveDrugsRow, error) {
	return nil, nil
}
func (m *mockDrugRepo) CreateDrugUnit(ctx context.Context, params sqlc.CreateDrugUnitParams) (sqlc.DrugUnit, error) {
	return sqlc.DrugUnit{}, nil
}
func (m *mockDrugRepo) ListDrugUnitsByDrug(ctx context.Context, params sqlc.ListDrugUnitsByDrugParams) ([]sqlc.DrugUnit, error) {
	return nil, nil
}
func (m *mockDrugRepo) CountDrugUnitsByDrug(ctx context.Context, drugID uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *mockDrugRepo) GetDrugUnit(ctx context.Context, id uuid.UUID) (sqlc.DrugUnit, error) {
	return sqlc.DrugUnit{}, nil
}
func (m *mockDrugRepo) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return nil, nil
}
func (m *mockDrugRepo) WithTx(tx pgx.Tx) *sqlc.Queries {
	return nil
}

func TestCreateDrug_Validation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := drugs.NewService(&mockDrugRepo{}, logger)
	userID := uuid.New()

	tests := []struct {
		name    string
		req     drugs.CreateRequest
		wantErr bool
	}{
		{
			name: "missing drug name",
			req: drugs.CreateRequest{
				Name:              "",
				ManufacturerID:    uuid.New().String(),
				BatchNumber:       "BATCH-01",
				ManufacturingDate: "2025-01-01",
				ExpiryDate:        "2026-01-01",
			},
			wantErr: true,
		},
		{
			name: "invalid manufacturer UUID",
			req: drugs.CreateRequest{
				Name:              "Paracetamol",
				ManufacturerID:    "not-a-uuid",
				BatchNumber:       "BATCH-01",
				ManufacturingDate: "2025-01-01",
				ExpiryDate:        "2026-01-01",
			},
			wantErr: true,
		},
		{
			name: "manufacturing date after expiry date",
			req: drugs.CreateRequest{
				Name:              "Paracetamol",
				ManufacturerID:    uuid.New().String(),
				BatchNumber:       "BATCH-01",
				ManufacturingDate: "2026-05-01",
				ExpiryDate:        "2025-05-01", // earlier than mfg
			},
			wantErr: true,
		},
		{
			name: "negative quantity",
			req: drugs.CreateRequest{
				Name:              "Paracetamol",
				ManufacturerID:    uuid.New().String(),
				BatchNumber:       "BATCH-01",
				ManufacturingDate: "2025-01-01",
				ExpiryDate:        "2026-01-01",
				Quantity:          -5,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.Create(context.Background(), tt.req, userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateUnits_QuantityBounds(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service := drugs.NewService(&mockDrugRepo{}, logger)
	drugID := uuid.New()

	// Zero quantity
	_, err := service.GenerateUnits(context.Background(), drugID, 0)
	if err == nil {
		t.Errorf("expected error for quantity 0")
	}

	// Negative quantity
	_, err = service.GenerateUnits(context.Background(), drugID, -10)
	if err == nil {
		t.Errorf("expected error for negative quantity")
	}

	// Excessive quantity (> 1000)
	_, err = service.GenerateUnits(context.Background(), drugID, 1001)
	if err == nil {
		t.Errorf("expected error for quantity > 1000")
	}
}
