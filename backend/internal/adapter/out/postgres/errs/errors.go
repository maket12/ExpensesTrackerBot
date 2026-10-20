package errs

import (
	"errors"

	pkgerrs "backend/pkg/errs"
)

var (
	ErrConnectionFailed = errors.New("failed to connect to postgresql database")
	ErrNotFound         = pkgerrs.ErrObjectNotFound
	ErrExecutionFailed  = errors.New("database query execution failed")
)

