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

func (uc *DeleteExpenseUC) validate(in dto.DeleteExpenseInput) error {
	if in.ID == uuid.Nil {
		return ucerrs.Wrap(
			ucerrs.ErrInvalidInput,
			pkgerrs.NewValueInvalidError("id"),
		)
	}
	return nil
}

func (uc *DeleteExpenseUC) Execute(ctx context.Context, in dto.DeleteExpenseInput) (*dto.DeleteExpenseOutput, error) {
	if err := uc.validate(in); err != nil {
		return nil, err
	}

	if in.EnforceUserID {
		existing, err := uc.repo.Get(ctx, in.ID)
		if err != nil {
			if errors.Is(err, pkgerrs.ErrObjectNotFound) {
				return nil, ucerrs.ErrExpenseNotFound
			}
			return nil, ucerrs.Wrap(ucerrs.ErrGetExpenseDB, err)
		}
		if existing.UserID() != in.UserID {
			return nil, ucerrs.ErrForbiddenAccess
		}
	}

	if err := uc.repo.Delete(ctx, in.ID); err != nil {
		if errors.Is(err, pkgerrs.ErrObjectNotFound) {
			return nil, ucerrs.ErrExpenseNotFound
		}
		return nil, ucerrs.Wrap(ucerrs.ErrDeleteExpenseDB, err)
	}

	return mapper.ToDeleteExpenseOutput(), nil
}
