package drugs

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/drug-verification/server/internal/database/sqlc"
)

// PostgresRepository implements Repository using PostgreSQL.
type PostgresRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

// NewPostgresRepository creates a new drug repository.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

func (r *PostgresRepository) CreateDrug(ctx context.Context, params sqlc.CreateDrugParams) (sqlc.Drug, error) {
	return r.queries.CreateDrug(ctx, params)
}

func (r *PostgresRepository) GetDrug(ctx context.Context, id uuid.UUID) (sqlc.GetDrugRow, error) {
	return r.queries.GetDrug(ctx, id)
}

func (r *PostgresRepository) ListDrugs(ctx context.Context, params sqlc.ListDrugsParams) ([]sqlc.ListDrugsRow, error) {
	return r.queries.ListDrugs(ctx, params)
}

func (r *PostgresRepository) CountDrugs(ctx context.Context, params sqlc.CountDrugsParams) (int64, error) {
	return r.queries.CountDrugs(ctx, params)
}

func (r *PostgresRepository) UpdateDrug(ctx context.Context, params sqlc.UpdateDrugParams) (sqlc.Drug, error) {
	return r.queries.UpdateDrug(ctx, params)
}

func (r *PostgresRepository) UpdateDrugStatus(ctx context.Context, params sqlc.UpdateDrugStatusParams) (sqlc.Drug, error) {
	return r.queries.UpdateDrugStatus(ctx, params)
}

func (r *PostgresRepository) GetDrugByManufacturerAndBatch(ctx context.Context, params sqlc.GetDrugByManufacturerAndBatchParams) (uuid.UUID, error) {
	return r.queries.GetDrugByManufacturerAndBatch(ctx, params)
}

func (r *PostgresRepository) GetExpiredActiveDrugs(ctx context.Context) ([]sqlc.GetExpiredActiveDrugsRow, error) {
	return r.queries.GetExpiredActiveDrugs(ctx)
}

func (r *PostgresRepository) CreateDrugUnit(ctx context.Context, params sqlc.CreateDrugUnitParams) (sqlc.DrugUnit, error) {
	return r.queries.CreateDrugUnit(ctx, params)
}

func (r *PostgresRepository) ListDrugUnitsByDrug(ctx context.Context, params sqlc.ListDrugUnitsByDrugParams) ([]sqlc.DrugUnit, error) {
	return r.queries.ListDrugUnitsByDrug(ctx, params)
}

func (r *PostgresRepository) CountDrugUnitsByDrug(ctx context.Context, drugID uuid.UUID) (int64, error) {
	return r.queries.CountDrugUnitsByDrug(ctx, drugID)
}

func (r *PostgresRepository) GetDrugUnit(ctx context.Context, id uuid.UUID) (sqlc.DrugUnit, error) {
	return r.queries.GetDrugUnit(ctx, id)
}

func (r *PostgresRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}

func (r *PostgresRepository) WithTx(tx pgx.Tx) *sqlc.Queries {
	return r.queries.WithTx(tx)
}
