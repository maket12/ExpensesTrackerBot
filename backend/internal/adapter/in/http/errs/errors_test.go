package errs

import (
	"errors"
	"net/http"
	"testing"

	apperrs "backend/internal/app/errs"
	"backend/internal/domain/models"
	pkgerrs "backend/pkg/errs"
)

func TestFromDomainOrAppError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode int
	}{
		{
			name:         "not found error",
			err:          pkgerrs.NewObjectNotFoundError("expense", "test-id"),
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "app not found error",
			err:          apperrs.ErrExpenseNotFound,
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "app forbidden error",
			err:          apperrs.ErrForbiddenAccess,
			expectedCode: http.StatusForbidden,
		},
		{
			name:         "invalid amount",
			err:          pkgerrs.NewValueInvalidError("amount"),
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid category",
			err:          pkgerrs.NewValueInvalidError("category"),
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "invalid user id",
			err:          pkgerrs.NewValueInvalidError("user_id"),
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "spent_at after created_at error",
			err:          models.ErrSpentAtAfterCreatedAt,
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
