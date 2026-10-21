package postgres

import (
	"context"

	"backend/internal/adapter/out/postgres/sqlc"
	pkgpostgres "backend/pkg/postgres"

	"github.com/jackc/pgx/v5/pgxpool"
)

type BaseRepository struct {
	q    *sqlc.Queries
	pool *pgxpool.Pool
}

func NewBaseRepository(pgClient *pkgpostgres.Client) BaseRepository {
	return BaseRepository{
		q:    sqlc.New(),
		pool: pgClient.Pool,
	}
}

func (b *BaseRepository) db(ctx context.Context) sqlc.DBTX {
	return b.pool
}
