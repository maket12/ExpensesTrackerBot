package errs

import (
	"errors"
	"fmt"
)

var ErrValueIsInvalid = errors.New("value is invalid")

type ValueInvalidError struct {
	ParamName string
	Reason    error
}

func NewValueInvalidErrorWithReason(paramName string, reason error) *ValueInvalidError {
	return &ValueInvalidError{
		ParamName: paramName,
		Reason:    reason,
	}
}

func NewValueInvalidError(paramName string) *ValueInvalidError {
	return &ValueInvalidError{
		ParamName: paramName,
	}
}

func (e *ValueInvalidError) Error() string {
	if e.Reason != nil {
		return fmt.Sprintf("%s: %s (reason: %v)",
			ErrValueIsInvalid, e.ParamName, e.Reason,
		)
	}
	return fmt.Sprintf("%s: %s", ErrValueIsInvalid, e.ParamName)
}

func (e *ValueInvalidError) Unwrap() error {
	return ErrValueIsInvalid
}
