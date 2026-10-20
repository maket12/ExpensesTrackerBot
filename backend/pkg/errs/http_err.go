package errs

import (
	"errors"
	"log/slog"
	"net/http"
)

var (
	ErrInvalidJSON       = errors.New("invalid json")
	ErrInvalidIdentifier = errors.New("invalid identifier format")
)

type OutErr struct {
	Code    int
	Message string
	Reason  error
	Level   slog.Level
}

func NewOutError(code int, msg string, reason error) *OutErr {
	return &OutErr{
		Code:    code,
		Message: msg,
		Reason:  reason,
		Level:   levelForCode(code),
	}
}

func levelForCode(code int) slog.Level {
	switch code {
	case http.StatusBadRequest,
		http.StatusNotFound,
		http.StatusUnprocessableEntity,
		http.StatusConflict:
		return slog.LevelWarn
	default:
		return slog.LevelError
	}
}
