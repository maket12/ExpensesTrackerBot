package usecases

import (
	"context"

	"backend/internal/app/dto"
	"backend/internal/app/mapper"
	"backend/internal/domain/models"
	"backend/internal/domain/port"
)

type CreateExpenseUC struct {
	repo port.ExpenseRepository
}

func NewCreateExpenseUC(repo port.ExpenseRepository) *CreateExpenseUC {
	return &CreateExpenseUC{
		repo: repo,
	}
}

func (uc *CreateExpenseUC) Execute(ctx context.Context, input dto.CreateExpenseInput) (dto.CreateExpenseOutput, error) {
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

	created, err := uc.repo.Create(ctx, expense)
	if err != nil {
		return dto.CreateExpenseOutput{}, err
	}

	return mapper.ToCreateExpenseOutput(created), nil
}
