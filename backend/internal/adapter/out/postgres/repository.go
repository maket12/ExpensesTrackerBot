package postgres

import (
	"context"
	"errors"

	"backend/internal/adapter/out/postgres/mapper"
	"backend/internal/adapter/out/postgres/sqlc"
	"backend/internal/domain/models"
	pkgerrs "backend/pkg/errs"
	pkgpostgres "backend/pkg/postgres"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	client  *pkgpostgres.Client
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewRepository(client *pkgpostgres.Client) *Repository {
	return &Repository{
		client:  client,
		pool:    client.Pool,
		queries: sqlc.New(client.Pool),
	}
}

func NewRepositoryWithPool(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

func (r *Repository) Create(ctx context.Context, expense *models.Expense) (*models.Expense, error) {
	params := mapper.ToCreateExpenseParams(expense)
	row, err := r.queries.CreateExpense(ctx, params)
	if err != nil {
		return nil, err
	}
	return mapper.ToDomainExpense(row)
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*models.Expense, error) {
	row, err := r.queries.GetExpenseByID(ctx, pkgpostgres.MapToPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pkgerrs.NewObjectNotFoundError("expense", id)
		}
		return nil, err
	}
	return mapper.ToDomainExpense(row)
}

func (r *Repository) Update(ctx context.Context, expense *models.Expense) (*models.Expense, error) {
	params := mapper.ToUpdateExpenseParams(expense)
	row, err := r.queries.UpdateExpense(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pkgerrs.NewObjectNotFoundError("expense", expense.ID())
		}
		return nil, err
	}
	return mapper.ToDomainExpense(row)
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteExpense(ctx, pkgpostgres.MapToPgUUID(id))
}

func (r *Repository) List(ctx context.Context, filter models.ExpenseFilter) ([]*models.Expense, error) {
	params := mapper.ToListExpensesFilteredParams(filter)
	rows, err := r.queries.ListExpensesFiltered(ctx, params)
	if err != nil {
		return nil, err
	}
	return mapper.ToDomainExpenses(rows)
}

func (r *Repository) GetLatestByUser(ctx context.Context, userID int64, limit int) ([]*models.Expense, error) {
	if limit <= 0 {
		limit = 5
	}
	rows, err := r.queries.GetLatestExpensesByUser(ctx, sqlc.GetLatestExpensesByUserParams{
		UserID: userID,
		Limit:  int32(limit),
	})
	if err != nil {
		return nil, err
	}
	return mapper.ToDomainExpenses(rows)
}

func (r *Repository) GetDistinctUserIDs(ctx context.Context) ([]int64, error) {
	return r.queries.GetDistinctUserIDs(ctx)
}

