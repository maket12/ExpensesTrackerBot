package errs

import (
	pkgerrs "backend/pkg/errs"
)

var (
	ErrEmptyToken   = pkgerrs.NewValueRequiredError("telegram_bot_token")
	ErrBotNotReady  = pkgerrs.NewValueInvalidError("telegram_bot_not_ready")
	ErrEmptyMessage = pkgerrs.NewValueRequiredError("telegram_message_text")
)

