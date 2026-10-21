package usecases

import (
	"context"
	"time"

	"backend/internal/app/dto"
	apperrs "backend/internal/app/errs"
	"backend/internal/app/mapper"
	"backend/internal/domain/models"
	"backend/internal/domain/port"
	pkgerrs "backend/pkg/errs"

	"github.com/google/uuid"
)

type ExpenseUseCase struct {
	repo port.ExpenseRepository
}

func NewExpenseUseCase(repo port.ExpenseRepository) *ExpenseUseCase {
	return &ExpenseUseCase{
		repo: repo,
	}
}

func (u *ExpenseUseCase) CreateExpense(ctx context.Context, input dto.CreateExpenseInput) (dto.CreateExpenseOutput, error) {
	expense, err := models.NewExpense(
		input.UserID,
		input.Amount,
		input.Category,
		input.SpentAt,
		input.Comment,
	)
	if err != nil {
		return dto.CreateExpenseOutput{}, err
	}

	created, err := u.repo.Create(ctx, expense)
	if err != nil {
		return dto.CreateExpenseOutput{}, err
	}

	return dto.CreateExpenseOutput{
		Expense: mapper.ToExpenseDTO(created),
	}, nil
}

func (u *ExpenseUseCase) GetExpense(ctx context.Context, input dto.GetExpenseInput) (dto.GetExpenseOutput, error) {
	id, err := uuid.Parse(input.ID)
	if err != nil {
		return dto.GetExpenseOutput{}, apperrs.NewInvalidInputError("id", pkgerrs.ErrInvalidIdentifier)
	}

	expense, err := u.repo.Get(ctx, id)
	if err != nil {
		return dto.GetExpenseOutput{}, err
	}

	return dto.GetExpenseOutput{
		Expense: mapper.ToExpenseDTO(expense),
	}, nil
}

func (u *ExpenseUseCase) UpdateExpense(ctx context.Context, input dto.UpdateExpenseInput) (dto.UpdateExpenseOutput, error) {
	id, err := uuid.Parse(input.ID)
	if err != nil {
		return dto.UpdateExpenseOutput{}, apperrs.NewInvalidInputError("id", pkgerrs.ErrInvalidIdentifier)
	}

	existing, err := u.repo.Get(ctx, id)
	if err != nil {
		return dto.UpdateExpenseOutput{}, err
	}

	if input.EnforceUserID && existing.UserID() != input.UserID {
		return dto.UpdateExpenseOutput{}, apperrs.ErrForbiddenAccess
	}

	var amountPtr *float64
	if input.Amount > 0 {
		amountPtr = &input.Amount
	}

	var catPtr *string
	if input.Category != "" {
		catPtr = &input.Category
	}

	var spentAtPtr *time.Time
	if !input.SpentAt.IsZero() {
		spentAtPtr = &input.SpentAt
	}

	var commPtr *string
	if input.Comment != "" {
		commPtr = &input.Comment
	}

	if err := existing.Update(amountPtr, catPtr, spentAtPtr, commPtr); err != nil {
		return dto.UpdateExpenseOutput{}, err
	}

	updated, err := u.repo.Update(ctx, existing)
	if err != nil {
		return dto.UpdateExpenseOutput{}, err
	}

	return dto.UpdateExpenseOutput{
		Expense: mapper.ToExpenseDTO(updated),
	}, nil
}

func (u *ExpenseUseCase) DeleteExpense(ctx context.Context, input dto.DeleteExpenseInput) (dto.DeleteExpenseOutput, error) {
	id, err := uuid.Parse(input.ID)
	if err != nil {
		return dto.DeleteExpenseOutput{}, apperrs.NewInvalidInputError("id", pkgerrs.ErrInvalidIdentifier)
	}

	if input.EnforceUserID {
		existing, err := u.repo.Get(ctx, id)
		if err != nil {
			return dto.DeleteExpenseOutput{}, err
		}
		if existing.UserID() != input.UserID {
			return dto.DeleteExpenseOutput{}, apperrs.ErrForbiddenAccess
		}
	}

	if err := u.repo.Delete(ctx, id); err != nil {
		return dto.DeleteExpenseOutput{}, err
	}

	return dto.DeleteExpenseOutput{Success: true}, nil
}

func (u *ExpenseUseCase) ListExpenses(ctx context.Context, input dto.ListExpensesInput) (dto.ListExpensesOutput, error) {
	filter, err := mapper.ToDomainFilter(input)
	if err != nil {
		return dto.ListExpensesOutput{}, err
	}

	expenses, err := u.repo.List(ctx, filter)
	if err != nil {
		return dto.ListExpensesOutput{}, err
	}

	return dto.ListExpensesOutput{
		Expenses: mapper.ToExpenseDTOList(expenses),
	}, nil
}

func (u *ExpenseUseCase) GetDistinctUsers(ctx context.Context, _ dto.GetDistinctUsersInput) (dto.GetDistinctUsersOutput, error) {
	userIDs, err := u.repo.ListUserIDS(ctx)
	if err != nil {
		return dto.GetDistinctUsersOutput{}, err
	}

	return dto.GetDistinctUsersOutput{
		UserIDs: userIDs,
	}, nil
}

func (u *ExpenseUseCase) GetLatestByUser(ctx context.Context, userID int64, limit int) ([]*models.Expense, error) {
	return u.repo.ListLatest(ctx, userID, limit)
}
