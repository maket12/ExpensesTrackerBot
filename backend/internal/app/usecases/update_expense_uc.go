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

type UpdateExpenseUC struct {
	repo port.ExpenseRepository
}

func NewUpdateExpenseUC(repo port.ExpenseRepository) *UpdateExpenseUC {
	return &UpdateExpenseUC{
		repo: repo,
	}
}

func (uc *UpdateExpenseUC) validate(in dto.UpdateExpenseInput) error {
	if _, err := uuid.Parse(in.ID); err != nil {
		return ucerrs.Wrap(
			ucerrs.ErrInvalidInput,
			pkgerrs.NewValueInvalidError("id"),
		)
	}
	return nil
}

func (uc *UpdateExpenseUC) Execute(ctx context.Context, in dto.UpdateExpenseInput) (*dto.UpdateExpenseOutput, error) {
	if err := uc.validate(in); err != nil {
		return nil, err
	}

	id, _ := uuid.Parse(in.ID)
	existing, err := uc.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, pkgerrs.ErrObjectNotFound) || errors.Is(err, ucerrs.ErrExpenseNotFound) {
			return nil, ucerrs.Wrap(ucerrs.ErrExpenseNotFound, err)
		}
		return nil, ucerrs.Wrap(ucerrs.ErrGetExpenseDB, err)
	}

	if in.EnforceUserID && existing.UserID() != in.UserID {
		return nil, ucerrs.Wrap(ucerrs.ErrForbiddenAccess, errors.New("user mismatch"))
	}

	if err := existing.Update(in.Amount, in.Category, in.SpentAt, in.Comment); err != nil {
		return nil, ucerrs.Wrap(ucerrs.ErrInvalidInput, err)
	}

	updated, err := uc.repo.Update(ctx, existing)
	if err != nil {
		if errors.Is(err, pkgerrs.ErrObjectNotFound) || errors.Is(err, ucerrs.ErrExpenseNotFound) {
			return nil, ucerrs.Wrap(ucerrs.ErrExpenseNotFound, err)
		}
		return nil, ucerrs.Wrap(ucerrs.ErrUpdateExpenseDB, err)
	}

	return mapper.ToUpdateExpenseOutput(updated), nil
}
