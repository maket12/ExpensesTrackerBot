package mapper

import (
	"errors"
	"net/http"
	"testing"

	apperrs "backend/internal/app/errs"
	"backend/internal/domain/models"
	pkgerrs "backend/pkg/errs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHttpError(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		assert.Nil(t, HttpError(nil))
	})

	t.Run("invalid json", func(t *testing.T) {
		out := HttpError(pkgerrs.ErrInvalidJSON)
		require.NotNil(t, out)
		assert.Equal(t, http.StatusBadRequest, out.Code)
		assert.Equal(t, pkgerrs.ErrInvalidJSON.Error(), out.Message)
	})

	t.Run("invalid identifier", func(t *testing.T) {
		out := HttpError(pkgerrs.ErrInvalidIdentifier)
		require.NotNil(t, out)
		assert.Equal(t, http.StatusBadRequest, out.Code)
	})

	t.Run("value is invalid", func(t *testing.T) {
		valErr := pkgerrs.NewValueInvalidError("id")
		out := HttpError(valErr)
		require.NotNil(t, out)
		assert.Equal(t, http.StatusBadRequest, out.Code)
	})

	t.Run("value is required", func(t *testing.T) {
		valErr := pkgerrs.NewValueRequiredError("amount")
		out := HttpError(valErr)
		require.NotNil(t, out)
		assert.Equal(t, http.StatusBadRequest, out.Code)
	})

	t.Run("spent at after created at", func(t *testing.T) {
		out := HttpError(models.ErrSpentAtAfterCreatedAt)
		require.NotNil(t, out)
		assert.Equal(t, http.StatusBadRequest, out.Code)
	})

	t.Run("wrapped invalid input", func(t *testing.T) {
		wrapped := apperrs.Wrap(apperrs.ErrInvalidInput, errors.New("field is missing"))
		out := HttpError(wrapped)
		require.NotNil(t, out)
		assert.Equal(t, http.StatusBadRequest, out.Code)
		assert.Contains(t, out.Message, "invalid input: field is missing")
	})

	t.Run("expense not found direct", func(t *testing.T) {
		out := HttpError(apperrs.ErrExpenseNotFound)
		require.NotNil(t, out)
		assert.Equal(t, http.StatusNotFound, out.Code)
	})

	t.Run("object not found direct", func(t *testing.T) {
		out := HttpError(pkgerrs.ErrObjectNotFound)
		require.NotNil(t, out)
		assert.Equal(t, http.StatusNotFound, out.Code)
	})

	t.Run("wrapped expense not found", func(t *testing.T) {
		wrapped := apperrs.Wrap(apperrs.ErrExpenseNotFound, errors.New("no row"))
		out := HttpError(wrapped)
		require.NotNil(t, out)
		assert.Equal(t, http.StatusNotFound, out.Code)
	})

	t.Run("forbidden access direct", func(t *testing.T) {
		out := HttpError(apperrs.ErrForbiddenAccess)
		require.NotNil(t, out)
		assert.Equal(t, http.StatusForbidden, out.Code)
	})

	t.Run("wrapped forbidden access", func(t *testing.T) {
		wrapped := apperrs.Wrap(apperrs.ErrForbiddenAccess, nil)
		out := HttpError(wrapped)
		require.NotNil(t, out)
		assert.Equal(t, http.StatusForbidden, out.Code)
	})

	t.Run("object already exists", func(t *testing.T) {
		out := HttpError(pkgerrs.ErrObjectAlreadyExists)
		require.NotNil(t, out)
		assert.Equal(t, http.StatusConflict, out.Code)
	})

	t.Run("wrapped db error", func(t *testing.T) {
		wrapped := apperrs.Wrap(apperrs.ErrCreateExpenseDB, errors.New("connection refused"))
		out := HttpError(wrapped)
		require.NotNil(t, out)
		assert.Equal(t, http.StatusInternalServerError, out.Code)
		assert.Equal(t, "internal error", out.Message)
		assert.Equal(t, errors.New("connection refused"), out.Reason)
	})

	t.Run("direct db error", func(t *testing.T) {
		out := HttpError(apperrs.ErrGetExpenseDB)
		require.NotNil(t, out)
		assert.Equal(t, http.StatusInternalServerError, out.Code)
		assert.Equal(t, "internal error", out.Message)
	})

	t.Run("unknown error", func(t *testing.T) {
		unknown := errors.New("something totally unknown")
		out := HttpError(unknown)
		require.NotNil(t, out)
		assert.Equal(t, http.StatusInternalServerError, out.Code)
		assert.Equal(t, "internal error", out.Message)
	})
}

func TestMapErrorToResponse(t *testing.T) {
	resp := MapErrorToResponse("something failed")
	assert.Equal(t, "something failed", resp.Error)
}
