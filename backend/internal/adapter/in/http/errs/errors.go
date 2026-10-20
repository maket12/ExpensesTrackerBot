package errs

import (
	"errors"
	"net/http"

	apperrs "backend/internal/app/errs"
	"backend/internal/domain/models"
	pkgerrs "backend/pkg/errs"
)

func FromDomainOrAppError(err error) *pkgerrs.OutErr {
	if errors.Is(err, apperrs.ErrExpenseNotFound) ||
		errors.Is(err, pkgerrs.ErrObjectNotFound) {
		return pkgerrs.NewOutError(http.StatusNotFound, err.Error(), err)
	}

	if errors.Is(err, apperrs.ErrForbiddenAccess) {
		return pkgerrs.NewOutError(http.StatusForbidden, err.Error(), err)
	}

	if errors.Is(err, pkgerrs.ErrNotAuthenticated) {
		return pkgerrs.NewOutError(http.StatusUnauthorized, err.Error(), err)
	}

	if errors.Is(err, pkgerrs.ErrObjectAlreadyExists) {
		return pkgerrs.NewOutError(http.StatusConflict, err.Error(), err)
	}

	if errors.Is(err, pkgerrs.ErrValueIsInvalid) ||
		errors.Is(err, pkgerrs.ErrValueIsRequired) ||
		errors.Is(err, pkgerrs.ErrInvalidJSON) ||
		errors.Is(err, pkgerrs.ErrInvalidIdentifier) ||
		errors.Is(err, models.ErrSpentAtAfterCreatedAt) ||
		errors.Is(err, apperrs.ErrInvalidInput) {
		return pkgerrs.NewOutError(http.StatusBadRequest, err.Error(), err)
	}

	return pkgerrs.NewOutError(http.StatusInternalServerError, "internal server error", err)
}
