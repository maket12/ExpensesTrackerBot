package errs

import (
	"errors"

	pkgerrs "backend/pkg/errs"
)

var (
	ErrEmptyAPIKey      = pkgerrs.NewValueRequiredError("llm_api_key")
	ErrEmptyResponse    = errors.New("llm returned empty response")
	ErrCompletionFailed = errors.New("llm completion failed")
)
