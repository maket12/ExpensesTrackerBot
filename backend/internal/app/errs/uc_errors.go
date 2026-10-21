package errs

import (
	"errors"

	pkgerrs "backend/pkg/errs"
)

/*
================ Invalidation & Conflicts ================
*/
var (
	ErrInvalidInput    = errors.New("invalid input")
	ErrForbiddenAccess = errors.New("access forbidden")
)

/*
================ Adapters failures ================
*/
var (
	ErrExpenseNotFound = pkgerrs.ErrObjectNotFound

	ErrAgentMaxIterations = errors.New("agent reached maximum iteration limit")
	ErrAgentFailed        = errors.New("agent failed to process message")
)

func NewInvalidInputError(paramName string, reason error) *pkgerrs.ValueInvalidError {
	if reason != nil {
		return pkgerrs.NewValueInvalidErrorWithReason(paramName, reason)
	}
	return pkgerrs.NewValueInvalidError(paramName)
}
