package usecases

import (
	"context"

	"backend/internal/app/dto"
	apperrs "backend/internal/app/errs"
	"backend/internal/app/mapper"
	"backend/internal/domain/port"
	pkgerrs "backend/pkg/errs"

	"github.com/google/uuid"
)

type DeleteExpenseUC struct {
	repo port.ExpenseRepository
}

func NewDeleteExpenseUC(repo port.ExpenseRepository) *DeleteExpenseUC {
	return &DeleteExpenseUC{
		repo: repo,
	}
}

func (uc *DeleteExpenseUC) Execute(ctx context.Context, input dto.DeleteExpenseInput) (dto.DeleteExpenseOutput, error) {
	id, err := uuid.Parse(input.ID)
	if err != nil {
		return dto.DeleteExpenseOutput{}, apperrs.NewInvalidInputError("id", pkgerrs.ErrInvalidIdentifier)
	}

	if input.EnforceUserID {
		existing, err := uc.repo.Get(ctx, id)
		if err != nil {
			return dto.DeleteExpenseOutput{}, err
		}
		if existing.UserID() != input.UserID {
			return dto.DeleteExpenseOutput{}, apperrs.ErrForbiddenAccess
		}
	}

	if err := uc.repo.Delete(ctx, id); err != nil {
		return dto.DeleteExpenseOutput{}, err
	}

	return mapper.ToDeleteExpenseOutput(), nil
}
