package usecases

import (
	"context"
	"errors"

	"backend/internal/app/dto"
	ucerrs "backend/internal/app/errs"
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

func (uc *GetExpenseUC) Execute(ctx context.Context, in dto.GetExpenseInput) (*dto.GetExpenseOutput, error) {
	id, err := uuid.Parse(in.ID)
	if err != nil {
		return nil, ucerrs.Wrap(ucerrs.ErrInvalidInput, err)
	}

	expense, err := uc.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, pkgerrs.ErrObjectNotFound) || errors.Is(err, ucerrs.ErrExpenseNotFound) {
			return nil, ucerrs.Wrap(ucerrs.ErrExpenseNotFound, err)
		}
		return nil, ucerrs.Wrap(ucerrs.ErrGetExpenseDB, err)
	}

	return mapper.ToGetExpenseOutput(expense), nil
}
