package manufacturers

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/drug-verification/server/internal/database/sqlc"
)

// PostgresRepository implements Repository using PostgreSQL.
type PostgresRepository struct {
	queries *sqlc.Queries
}

// NewPostgresRepository creates a new manufacturer repository.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		queries: sqlc.New(pool),
	}
}

func (r *PostgresRepository) CreateManufacturer(ctx context.Context, params sqlc.CreateManufacturerParams) (sqlc.Manufacturer, error) {
	return r.queries.CreateManufacturer(ctx, params)
}

func (r *PostgresRepository) GetManufacturer(ctx context.Context, id uuid.UUID) (sqlc.Manufacturer, error) {
	return r.queries.GetManufacturer(ctx, id)
}

func (r *PostgresRepository) ListManufacturers(ctx context.Context, params sqlc.ListManufacturersParams) ([]sqlc.Manufacturer, error) {
	return r.queries.ListManufacturers(ctx, params)
}

func (r *PostgresRepository) CountManufacturers(ctx context.Context) (int64, error) {
	return r.queries.CountManufacturers(ctx)
}

func (r *PostgresRepository) UpdateManufacturer(ctx context.Context, params sqlc.UpdateManufacturerParams) (sqlc.Manufacturer, error) {
	return r.queries.UpdateManufacturer(ctx, params)
}

func (r *PostgresRepository) DeleteManufacturer(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteManufacturer(ctx, id)
}
