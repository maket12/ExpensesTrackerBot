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

type UpdateExpenseUC struct {
	repo port.ExpenseRepository
}

func NewUpdateExpenseUC(repo port.ExpenseRepository) *UpdateExpenseUC {
	return &UpdateExpenseUC{
		repo: repo,
	}
}

func (uc *UpdateExpenseUC) Execute(ctx context.Context, input dto.UpdateExpenseInput) (dto.UpdateExpenseOutput, error) {
	id, err := uuid.Parse(input.ID)
	if err != nil {
		return dto.UpdateExpenseOutput{}, apperrs.NewInvalidInputError("id", pkgerrs.ErrInvalidIdentifier)
	}

	existing, err := uc.repo.Get(ctx, id)
	if err != nil {
		return dto.UpdateExpenseOutput{}, err
	}

	if input.EnforceUserID && existing.UserID() != input.UserID {
		return dto.UpdateExpenseOutput{}, apperrs.ErrForbiddenAccess
	}

	if err := existing.Update(input.Amount, input.Category, input.SpentAt, input.Comment); err != nil {
		return dto.UpdateExpenseOutput{}, err
	}

	updated, err := uc.repo.Update(ctx, existing)
	if err != nil {
		return dto.UpdateExpenseOutput{}, err
	}

	return mapper.ToUpdateExpenseOutput(updated), nil
}
