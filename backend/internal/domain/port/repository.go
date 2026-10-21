package port

import (
	"context"

	"backend/internal/domain/models"

	"github.com/google/uuid"
)

type ExpenseRepository interface {
	Create(ctx context.Context, expense *models.Expense) (*models.Expense, error)
	Get(ctx context.Context, id uuid.UUID) (*models.Expense, error)
	Update(ctx context.Context, expense *models.Expense) (*models.Expense, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter ExpenseFilter) ([]*models.Expense, error)
	ListLatest(ctx context.Context, userID int64, limit int) ([]*models.Expense, error)
	ListUserIDS(ctx context.Context) ([]int64, error)
}
