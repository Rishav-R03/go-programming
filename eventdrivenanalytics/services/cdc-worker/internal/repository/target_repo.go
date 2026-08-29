package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TargetRepository struct {
	pool *pgxpool.Pool
}

func NewTargetRepository(pool *pgxpool.Pool) *TargetRepository {
	return &TargetRepository{pool: pool}
}

func (r *TargetRepository) InsertFact(ctx context.Context, query string, args ...any) error {
	_, err := r.pool.Exec(ctx, query, args...)
	return err
}
