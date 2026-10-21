package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"backend/internal/app/dto"
	apperrs "backend/internal/app/errs"
	"backend/internal/domain/models"
	"backend/internal/domain/port/mocks"
	pkgerrs "backend/pkg/errs"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDeleteExpenseUC_Execute_Success_WithoutEnforceUserID(t *testing.T) {
	repo := mocks.NewMockExpenseRepository(t)
	uc := NewDeleteExpenseUC(repo)

	id := uuid.New()
	repo.EXPECT().
		Delete(mock.Anything, id).
		Return(nil).
		Once()

	out, err := uc.Execute(context.Background(), dto.DeleteExpenseInput{
		ID:            id.String(),
		EnforceUserID: false,
	})

	require.NoError(t, err)
	require.NotNil(t, out)
	assert.True(t, out.Success)
}

func TestDeleteExpenseUC_Execute_Success_WithEnforceUserID(t *testing.T) {
	repo := mocks.NewMockExpenseRepository(t)
	uc := NewDeleteExpenseUC(repo)

	e, err := models.NewExpense(100, 300, "еда", time.Now().UTC().Add(-time.Hour), "обед")
	require.NoError(t, err)

	repo.EXPECT().
		Get(mock.Anything, e.ID()).
		Return(e, nil).
		Once()

	repo.EXPECT().
		Delete(mock.Anything, e.ID()).
		Return(nil).
		Once()

	out, err := uc.Execute(context.Background(), dto.DeleteExpenseInput{
		ID:            e.ID().String(),
		UserID:        100,
		EnforceUserID: true,
	})

	require.NoError(t, err)
	require.NotNil(t, out)
	assert.True(t, out.Success)
}

func TestDeleteExpenseUC_Execute_InvalidID(t *testing.T) {
	repo := mocks.NewMockExpenseRepository(t)
	uc := NewDeleteExpenseUC(repo)

	out, err := uc.Execute(context.Background(), dto.DeleteExpenseInput{ID: "invalid-uuid"})
	assert.Nil(t, out)
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrs.ErrInvalidInput))
}

func TestDeleteExpenseUC_Execute_EnforceUserID_GetNotFound(t *testing.T) {
	repo := mocks.NewMockExpenseRepository(t)
	uc := NewDeleteExpenseUC(repo)

	id := uuid.New()
	repo.EXPECT().
		Get(mock.Anything, id).
		Return(nil, pkgerrs.NewObjectNotFoundError("expense", id)).
		Once()

	out, err := uc.Execute(context.Background(), dto.DeleteExpenseInput{
		ID:            id.String(),
		UserID:        100,
		EnforceUserID: true,
	})

	assert.Nil(t, out)
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrs.ErrExpenseNotFound))
}

func TestDeleteExpenseUC_Execute_EnforceUserID_GetRepoError(t *testing.T) {
	repo := mocks.NewMockExpenseRepository(t)
	uc := NewDeleteExpenseUC(repo)

	id := uuid.New()
	repo.EXPECT().
		Get(mock.Anything, id).
		Return(nil, errors.New("db error")).
		Once()

	out, err := uc.Execute(context.Background(), dto.DeleteExpenseInput{
		ID:            id.String(),
		UserID:        100,
		EnforceUserID: true,
	})

	assert.Nil(t, out)
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrs.ErrGetExpenseDB))
}

func TestDeleteExpenseUC_Execute_EnforceUserID_Forbidden(t *testing.T) {
	repo := mocks.NewMockExpenseRepository(t)
	uc := NewDeleteExpenseUC(repo)

	e, err := models.NewExpense(100, 300, "еда", time.Now().UTC().Add(-time.Hour), "обед")
	require.NoError(t, err)

	repo.EXPECT().
		Get(mock.Anything, e.ID()).
		Return(e, nil).
		Once()

	out, err := uc.Execute(context.Background(), dto.DeleteExpenseInput{
		ID:            e.ID().String(),
		UserID:        200,
		EnforceUserID: true,
	})

	assert.Nil(t, out)
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrs.ErrForbiddenAccess))
}

func TestDeleteExpenseUC_Execute_DeleteNotFound(t *testing.T) {
	repo := mocks.NewMockExpenseRepository(t)
	uc := NewDeleteExpenseUC(repo)

	id := uuid.New()
	repo.EXPECT().
		Delete(mock.Anything, id).
		Return(pkgerrs.NewObjectNotFoundError("expense", id)).
		Once()

	out, err := uc.Execute(context.Background(), dto.DeleteExpenseInput{
		ID:            id.String(),
		EnforceUserID: false,
	})

	assert.Nil(t, out)
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrs.ErrExpenseNotFound))
}

func TestDeleteExpenseUC_Execute_DeleteRepoError(t *testing.T) {
	repo := mocks.NewMockExpenseRepository(t)
	uc := NewDeleteExpenseUC(repo)

	id := uuid.New()
	repo.EXPECT().
		Delete(mock.Anything, id).
		Return(errors.New("db delete error")).
		Once()

	out, err := uc.Execute(context.Background(), dto.DeleteExpenseInput{
		ID:            id.String(),
		EnforceUserID: false,
	})

	assert.Nil(t, out)
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrs.ErrDeleteExpenseDB))
}
