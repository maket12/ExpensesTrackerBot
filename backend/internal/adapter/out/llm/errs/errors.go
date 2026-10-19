package errs

import "errors"

var (
	ErrEmptyAPIKey      = errors.New("llm api key is required")
	ErrEmptyResponse    = errors.New("llm returned empty response")
	ErrCompletionFailed = errors.New("llm completion failed")
)
