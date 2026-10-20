package errs

import (
	"errors"
	"fmt"
)

var ErrObjectNotFound = errors.New("object not found")

type ObjectNotFoundError struct {
	ParamName string
	ID        any
	Reason    error
}

func NewObjectNotFoundWithReasonError(paramName string, id string, reason error) *ObjectNotFoundError {
	return &ObjectNotFoundError{
		ParamName: paramName,
		ID:        id,
		Reason:    reason,
	}
}

func NewObjectNotFoundError(paramName string, id any) *ObjectNotFoundError {
	return &ObjectNotFoundError{
		ParamName: paramName,
		ID:        id,
	}
}

func (e *ObjectNotFoundError) Error() string {
	if e.Reason != nil {
		return fmt.Sprintf("%s: param is: %s, ID is: %s (reason: %v)",
			ErrObjectNotFound, e.ParamName, e.ID, e.Reason,
		)
	}
	return fmt.Sprintf("%s: %s", ErrObjectNotFound, e.ID)
}

func (e *ObjectNotFoundError) Unwrap() error {
	return ErrObjectNotFound
}
