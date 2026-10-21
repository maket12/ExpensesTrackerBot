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

type GetExpenseUC struct {
	repo port.ExpenseRepository
}

func NewGetExpenseUC(repo port.ExpenseRepository) *GetExpenseUC {
	return &GetExpenseUC{
		repo: repo,
	}
}

func (uc *GetExpenseUC) Execute(ctx context.Context, input dto.GetExpenseInput) (dto.GetExpenseOutput, error) {
	id, err := uuid.Parse(input.ID)
	if err != nil {
		return dto.GetExpenseOutput{}, apperrs.NewInvalidInputError("id", pkgerrs.ErrInvalidIdentifier)
	}

	expense, err := uc.repo.Get(ctx, id)
	if err != nil {
		return dto.GetExpenseOutput{}, err
	}

	return mapper.ToGetExpenseOutput(expense), nil
}
