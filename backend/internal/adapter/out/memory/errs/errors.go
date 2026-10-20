package errs

import (
	pkgerrs "backend/pkg/errs"
)

var (
	ErrEmptyMessage = pkgerrs.NewValueRequiredError("chat_message_content")
)
