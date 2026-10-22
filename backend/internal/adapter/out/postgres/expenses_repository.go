package postgres

import (
	"context"
	"errors"

	"backend/internal/adapter/out/postgres/mapper"
	"backend/internal/domain/models"
	"backend/internal/domain/port"
	pkgerrs "backend/pkg/errs"
	pkgpostgres "backend/pkg/postgres"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ExpensesRepository struct{ BaseRepository }

func NewExpensesRepository(client *pkgpostgres.Client) *ExpensesRepository {
	return &ExpensesRepository{
		BaseRepository: NewBaseRepository(client),
	}
}

func (r *ExpensesRepository) Create(ctx context.Context, expense *models.Expense) (*models.Expense, error) {
	params := mapper.ToCreateExpenseParams(expense)
	row, err := r.q.CreateExpense(ctx, r.db(ctx), params)
	if err != nil {
		return nil, err
	}
	return mapper.ToDomainExpense(row)
}

func (r *ExpensesRepository) Get(ctx context.Context, id uuid.UUID) (*models.Expense, error) {
	row, err := r.q.GetExpenseByID(ctx, r.db(ctx), pkgpostgres.MapToPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pkgerrs.NewObjectNotFoundError("expense", id)
		}
		return nil, err
	}
	return mapper.ToDomainExpense(row)
}

func (r *ExpensesRepository) Update(ctx context.Context, expense *models.Expense) (*models.Expense, error) {
	params := mapper.ToUpdateExpenseParams(expense)
	row, err := r.q.UpdateExpense(ctx, r.db(ctx), params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pkgerrs.NewObjectNotFoundError("expense", expense.ID())
		}
		return nil, err
	}
	return mapper.ToDomainExpense(row)
}

func (r *ExpensesRepository) Delete(ctx context.Context, id uuid.UUID) error {
	rows, err := r.q.DeleteExpense(ctx, r.db(ctx), pkgpostgres.MapToPgUUID(id))
	if err != nil {
		return err
	}
	if rows == 0 {
		return pkgerrs.NewObjectNotFoundError("expense", id)
	}
	return nil
}

func (r *ExpensesRepository) List(ctx context.Context, filter port.ExpenseFilter) ([]*models.Expense, error) {
	params := mapper.ToListExpensesFilteredParams(filter)
	rows, err := r.q.ListExpensesFiltered(ctx, r.db(ctx), params)
	if err != nil {
		return nil, err
	}
	return mapper.ToDomainExpenses(rows)
}

func (r *ExpensesRepository) ListLatest(ctx context.Context, userID int64, limit int) ([]*models.Expense, error) {
	params := mapper.ToListLatestExpensesParams(userID, limit)
	rows, err := r.q.GetLatestExpensesByUser(ctx, r.db(ctx), params)
	if err != nil {
		return nil, err
	}
	return mapper.ToDomainExpenses(rows)
}

func (r *ExpensesRepository) ListUserIDS(ctx context.Context) ([]int64, error) {
	return r.q.GetDistinctUserIDs(ctx, r.db(ctx))
}
