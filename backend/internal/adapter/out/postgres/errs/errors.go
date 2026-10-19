package errs

import "errors"

var (
	ErrConnectionFailed = errors.New("failed to connect to postgresql database")
	ErrNotFound         = errors.New("record not found in postgresql")
	ErrExecutionFailed  = errors.New("database query execution failed")
)
