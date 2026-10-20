package errs

import (
	"errors"
	"fmt"
)

var ErrValueIsRequired = errors.New("value is required")

type ValueRequiredError struct {
	ParamName string
	Reason    error
}

func NewValueRequiredErrorWithReason(paramName string, reason error) *ValueRequiredError {
	return &ValueRequiredError{
		ParamName: paramName,
		Reason:    reason,
	}
}

func NewValueRequiredError(paramName string) *ValueRequiredError {
	return &ValueRequiredError{
		ParamName: paramName,
	}
}

func (e *ValueRequiredError) Error() string {
	if e.Reason != nil {
		return fmt.Sprintf("%s: %s (reason: %v)",
			ErrValueIsRequired, e.ParamName, e.Reason,
		)
	}
	return fmt.Sprintf("%s: %s", ErrValueIsRequired, e.ParamName)
}

func (e *ValueRequiredError) Unwrap() error {
	return ErrValueIsRequired
}
