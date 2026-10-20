package errs

import (
	"errors"

	pkgerrs "backend/pkg/errs"
)

var (
	ErrExpenseNotFound    = pkgerrs.ErrObjectNotFound
	ErrForbiddenAccess    = errors.New("access forbidden")
	ErrAgentMaxIterations = errors.New("agent reached maximum iteration limit")
	ErrAgentFailed        = errors.New("agent failed to process message")
	ErrInvalidInput       = pkgerrs.ErrValueIsInvalid
)

func NewInvalidInputError(paramName string, reason error) *pkgerrs.ValueInvalidError {
	if reason != nil {
		return pkgerrs.NewValueInvalidErrorWithReason(paramName, reason)
	}
	return pkgerrs.NewValueInvalidError(paramName)
}

