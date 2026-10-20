package errs

import (
	"errors"
	"fmt"
)

var ErrObjectAlreadyExists = errors.New("object already exists")

type ObjectAlreadyExistsError struct {
	ParamName string
	Reason    error
}

func NewObjectAlreadyExistsErrorWithReason(paramName string, reason error) *ObjectAlreadyExistsError {
	return &ObjectAlreadyExistsError{
		ParamName: paramName,
		Reason:    reason,
	}
}

func NewObjectAlreadyExistsError(paramName string) *ObjectAlreadyExistsError {
	return &ObjectAlreadyExistsError{
		ParamName: paramName,
	}
}

func (e *ObjectAlreadyExistsError) Error() string {
	if e.Reason != nil {
		return fmt.Sprintf("%s: %s (reason: %v)",
			ErrObjectAlreadyExists, e.ParamName, e.Reason,
		)
	}
	return fmt.Sprintf("%s: %s", ErrObjectAlreadyExists, e.ParamName)
}

func (e *ObjectAlreadyExistsError) Unwrap() error {
	return ErrObjectAlreadyExists
}
