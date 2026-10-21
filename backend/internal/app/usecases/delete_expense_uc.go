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

type DeleteExpenseUC struct {
	repo port.ExpenseRepository
}

func NewDeleteExpenseUC(repo port.ExpenseRepository) *DeleteExpenseUC {
	return &DeleteExpenseUC{
		repo: repo,
	}
}

func (uc *DeleteExpenseUC) Execute(ctx context.Context, in dto.DeleteExpenseInput) (*dto.DeleteExpenseOutput, error) {
	id, err := uuid.Parse(in.ID)
	if err != nil {
		return nil, ucerrs.Wrap(ucerrs.ErrInvalidInput, err)
	}

	if in.EnforceUserID {
		existing, err := uc.repo.Get(ctx, id)
		if err != nil {
			if errors.Is(err, pkgerrs.ErrObjectNotFound) || errors.Is(err, ucerrs.ErrExpenseNotFound) {
				return nil, ucerrs.Wrap(ucerrs.ErrExpenseNotFound, err)
			}
			return nil, ucerrs.Wrap(ucerrs.ErrGetExpenseDB, err)
		}
		if existing.UserID() != in.UserID {
			return nil, ucerrs.Wrap(ucerrs.ErrForbiddenAccess, errors.New("user mismatch"))
		}
	}

	if err := uc.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, pkgerrs.ErrObjectNotFound) || errors.Is(err, ucerrs.ErrExpenseNotFound) {
			return nil, ucerrs.Wrap(ucerrs.ErrExpenseNotFound, err)
		}
		return nil, ucerrs.Wrap(ucerrs.ErrDeleteExpenseDB, err)
	}

	return mapper.ToDeleteExpenseOutput(), nil
}
