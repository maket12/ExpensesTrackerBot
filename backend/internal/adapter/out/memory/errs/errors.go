package errs

import "errors"

var (
	ErrEmptyMessage = errors.New("chat message content cannot be empty")
)
