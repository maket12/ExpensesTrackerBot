package usecases

import (
	"context"

	"backend/internal/app/dto"
	ucerrs "backend/internal/app/errs"
	"backend/internal/app/mapper"
	"backend/internal/domain/port"
)

type ListExpensesUC struct {
	repo port.ExpenseRepository
}

func NewListExpensesUC(repo port.ExpenseRepository) *ListExpensesUC {
	return &ListExpensesUC{repo: repo}
}

func (uc *ListExpensesUC) Execute(ctx context.Context, in dto.ListExpensesInput) (*dto.ListExpensesOutput, error) {
	filter, err := mapper.ToDomainFilter(in)
	if err != nil {
		return nil, ucerrs.Wrap(ucerrs.ErrInvalidInput, err)
	}

	expenses, err := uc.repo.List(ctx, filter)
	if err != nil {
		return nil, ucerrs.Wrap(ucerrs.ErrListExpensesDB, err)
	}

	return mapper.ToListExpensesOutput(expenses), nil
}
