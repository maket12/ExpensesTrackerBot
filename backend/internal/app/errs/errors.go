package errs

import "errors"

var (
	ErrExpenseNotFound    = errors.New("expense not found")
	ErrForbiddenAccess    = errors.New("access forbidden")
	ErrAgentMaxIterations = errors.New("agent reached maximum iteration limit")
	ErrAgentFailed        = errors.New("agent failed to process message")
	ErrInvalidInput       = errors.New("invalid use case input")
)
