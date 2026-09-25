package auth

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

// NewPostgresRepository creates a new auth repository.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		queries: sqlc.New(pool),
	}
}

func (r *PostgresRepository) CreateUser(ctx context.Context, params sqlc.CreateUserParams) (sqlc.CreateUserRow, error) {
	return r.queries.CreateUser(ctx, params)
}

func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (sqlc.User, error) {
	return r.queries.GetUserByEmail(ctx, email)
}

func (r *PostgresRepository) GetUserByID(ctx context.Context, id uuid.UUID) (sqlc.GetUserByIDRow, error) {
	return r.queries.GetUserByID(ctx, id)
}
