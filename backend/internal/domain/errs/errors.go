package errs

import (
	"errors"

	pkgerrs "backend/pkg/errs"
)

var (
	ErrInvalidAmount   = pkgerrs.NewValueInvalidErrorWithReason("amount", errors.New("amount must be greater than zero"))
	ErrInvalidCategory = pkgerrs.NewValueInvalidErrorWithReason("category", errors.New("invalid category"))
	ErrInvalidUserID   = pkgerrs.NewValueInvalidErrorWithReason("user_id", errors.New("user id must be greater than zero"))
	ErrInvalidSpentAt  = pkgerrs.NewValueInvalidErrorWithReason("spent_at", errors.New("spent_at cannot be zero"))
	ErrExpenseNotFound = pkgerrs.ErrObjectNotFound
	ErrForbiddenAccess = errors.New("access to expense is forbidden")
)

func NewExpenseNotFoundError(id any) *pkgerrs.ObjectNotFoundError {
	return pkgerrs.NewObjectNotFoundError("expense", id)
}

