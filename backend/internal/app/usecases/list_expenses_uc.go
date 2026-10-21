package usecases

import (
	"context"

	"backend/internal/app/dto"
	"backend/internal/app/mapper"
	"backend/internal/domain/port"
)

type ListExpensesUC struct {
	repo port.ExpenseRepository
}

func NewListExpensesUC(repo port.ExpenseRepository) *ListExpensesUC {
	return &ListExpensesUC{
		repo: repo,
	}
}

func (uc *ListExpensesUC) Execute(ctx context.Context, input dto.ListExpensesInput) (dto.ListExpensesOutput, error) {
	filter, err := mapper.ToDomainFilter(input)
	if err != nil {
		return dto.ListExpensesOutput{}, err
	}

	expenses, err := uc.repo.List(ctx, filter)
	if err != nil {
		return dto.ListExpensesOutput{}, err
	}

	return mapper.ToListExpensesOutput(expenses), nil
}
