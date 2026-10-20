package errs

import (
	"errors"
	"net/http"
	"testing"

	apperrs "backend/internal/app/errs"
	domainerrs "backend/internal/domain/errs"
	pkgerrs "backend/pkg/errs"
)

func TestFromDomainOrAppError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode int
	}{
		{
			name:         "domain not found error",
			err:          domainerrs.NewExpenseNotFoundError("test-id"),
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "app not found error",
			err:          apperrs.ErrExpenseNotFound,
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "forbidden error",
			err:          domainerrs.ErrForbiddenAccess,
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "app forbidden error",
			err:          apperrs.ErrForbiddenAccess,
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "invalid amount",
			err:          domainerrs.ErrInvalidAmount,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid category",
			err:          domainerrs.ErrInvalidCategory,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid user id",
			err:          domainerrs.ErrInvalidUserID,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid input",
			err:          apperrs.ErrInvalidInput,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid identifier",
			err:          pkgerrs.ErrInvalidIdentifier,
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "unauthenticated",
			err:          pkgerrs.ErrNotAuthenticated,
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "object already exists",
			err:          pkgerrs.ErrObjectAlreadyExists,
			expectedCode: http.StatusConflict,
		},
		{
			name:         "unknown error",
			err:          errors.New("random unexpected error"),
			expectedCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := FromDomainOrAppError(tt.err)
			if out == nil {
				t.Fatalf("expected non-nil OutErr")
			}
			if out.Code != tt.expectedCode {
				t.Errorf("expected status code %d, got %d", tt.expectedCode, out.Code)
			}
		})
	}
}
