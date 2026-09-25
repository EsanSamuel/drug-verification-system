package verification

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/drug-verification/server/internal/database/sqlc"
)

// PostgresRepository implements Repository using PostgreSQL.
type PostgresRepository struct {
	queries *sqlc.Queries
}

// NewPostgresRepository creates a new verification repository.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		queries: sqlc.New(pool),
	}
}

func (r *PostgresRepository) GetDrugUnitBySerial(ctx context.Context, serial string) (sqlc.GetDrugUnitBySerialRow, error) {
	return r.queries.GetDrugUnitBySerial(ctx, serial)
}

func (r *PostgresRepository) CreateVerificationLog(ctx context.Context, params sqlc.CreateVerificationLogParams) (sqlc.VerificationLog, error) {
	return r.queries.CreateVerificationLog(ctx, params)
}

func (r *PostgresRepository) ListVerificationLogs(ctx context.Context, params sqlc.ListVerificationLogsParams) ([]sqlc.ListVerificationLogsRow, error) {
	return r.queries.ListVerificationLogs(ctx, params)
}

func (r *PostgresRepository) CountVerificationLogs(ctx context.Context, params sqlc.CountVerificationLogsParams) (int64, error) {
	return r.queries.CountVerificationLogs(ctx, params)
}

func (r *PostgresRepository) ListVerificationLogsByDrug(ctx context.Context, params sqlc.ListVerificationLogsByDrugParams) ([]sqlc.ListVerificationLogsByDrugRow, error) {
	return r.queries.ListVerificationLogsByDrug(ctx, params)
}

func (r *PostgresRepository) CountVerificationLogsByDrug(ctx context.Context, params sqlc.CountVerificationLogsByDrugParams) (int64, error) {
	return r.queries.CountVerificationLogsByDrug(ctx, params)
}
