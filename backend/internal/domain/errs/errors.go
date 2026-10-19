package errs

import "errors"

var (
	ErrInvalidAmount   = errors.New("amount must be greater than zero")
	ErrInvalidCategory = errors.New("invalid category")
	ErrInvalidUserID   = errors.New("invalid user id")
	ErrInvalidSpentAt  = errors.New("invalid spent_at timestamp")
	ErrExpenseNotFound = errors.New("expense not found")
	ErrForbiddenAccess = errors.New("access to expense is forbidden")
)
