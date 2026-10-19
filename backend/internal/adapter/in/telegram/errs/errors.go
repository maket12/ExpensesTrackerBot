package errs

import "errors"

var (
	ErrEmptyToken   = errors.New("telegram bot token is empty")
	ErrBotNotReady  = errors.New("telegram bot is not initialized")
	ErrEmptyMessage = errors.New("telegram message text is empty")
)
