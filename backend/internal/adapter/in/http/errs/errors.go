package errs

import (
	"errors"
	"net/http"

	apperrs "backend/internal/app/errs"
	domainerrs "backend/internal/domain/errs"
)

type HTTPError struct {
	Code    int    `json:"code"`
	Message string `json:"error"`
}

func FromDomainOrAppError(err error) HTTPError {
	if errors.Is(err, domainerrs.ErrExpenseNotFound) || errors.Is(err, apperrs.ErrExpenseNotFound) {
		return HTTPError{Code: http.StatusNotFound, Message: err.Error()}
	}

	if errors.Is(err, domainerrs.ErrForbiddenAccess) || errors.Is(err, apperrs.ErrForbiddenAccess) {
		return HTTPError{Code: http.StatusForbidden, Message: err.Error()}
	}

	if errors.Is(err, domainerrs.ErrInvalidAmount) ||
		errors.Is(err, domainerrs.ErrInvalidCategory) ||
		errors.Is(err, domainerrs.ErrInvalidUserID) ||
		errors.Is(err, domainerrs.ErrInvalidSpentAt) ||
		errors.Is(err, apperrs.ErrInvalidInput) {
		return HTTPError{Code: http.StatusBadRequest, Message: err.Error()}
	}

	return HTTPError{Code: http.StatusInternalServerError, Message: "internal server error"}
}
