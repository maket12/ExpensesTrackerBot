package errs

import (
	"errors"
	"fmt"
)

var ErrNotAuthenticated = errors.New("you must be authenticated to make this request")

type NotAuthenticatedError struct {
	Reason error
}

func NewNotAuthenticatedErrorWithReason(reason error) *NotAuthenticatedError {
	return &NotAuthenticatedError{Reason: reason}
}

func NewNotAuthenticatedError() *NotAuthenticatedError {
	return &NotAuthenticatedError{}
}

func (e *NotAuthenticatedError) Error() string {
	if e.Reason != nil {
		return fmt.Sprintf("%s (reason: %v)",
			ErrNotAuthenticated, e.Reason,
		)
	}
	return ErrNotAuthenticated.Error()
}

func (e *NotAuthenticatedError) Unwrap() error {
	return ErrNotAuthenticated
}
