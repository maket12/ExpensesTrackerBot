package usecases

import (
	ucerrs "backend/internal/app/errs"
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

func (uc *CreateExpenseUC) Execute(ctx context.Context, in dto.CreateExpenseInput) (*dto.CreateExpenseOutput, error) {
	expense, err := models.NewExpense(
		in.UserID,
		in.Amount,
		in.Category,
		in.SpentAt,
		in.Comment,
	)
	if err != nil {
		return nil, ucerrs.Wrap(ucerrs.ErrInvalidInput, err)
	}

	created, err := uc.repo.Create(ctx, expense)
	if err != nil {
		return nil, ucerrs.Wrap(ucerrs.ErrCreateExpenseDB, err)
	}

	return mapper.ToCreateExpenseOutput(created), nil
}
