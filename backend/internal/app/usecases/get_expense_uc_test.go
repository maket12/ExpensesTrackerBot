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

func TestGetExpenseUC_Execute_Success(t *testing.T) {
	repo := mocks.NewMockExpenseRepository(t)
	uc := NewGetExpenseUC(repo)

	e, err := models.NewExpense(100, 150, "еда", time.Now().UTC().Add(-time.Minute), "кофе")
	require.NoError(t, err)

	repo.EXPECT().
		Get(mock.Anything, e.ID()).
		Return(e, nil).
		Once()

	out, err := uc.Execute(context.Background(), dto.GetExpenseInput{ID: e.ID()})
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, e.ID(), out.Expense.ID)
	assert.Equal(t, int64(100), out.Expense.UserID)
	assert.Equal(t, 150.0, out.Expense.Amount)
	assert.Equal(t, "еда", out.Expense.Category)
}

func TestGetExpenseUC_Execute_InvalidID(t *testing.T) {
	repo := mocks.NewMockExpenseRepository(t)
	uc := NewGetExpenseUC(repo)

	out, err := uc.Execute(context.Background(), dto.GetExpenseInput{ID: uuid.Nil})
	assert.Nil(t, out)
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrs.ErrInvalidInput))
}

func TestGetExpenseUC_Execute_NotFound(t *testing.T) {
	repo := mocks.NewMockExpenseRepository(t)
	uc := NewGetExpenseUC(repo)

	id := uuid.New()
	repo.EXPECT().
		Get(mock.Anything, id).
		Return(nil, pkgerrs.NewObjectNotFoundError("expense", id)).
		Once()

	out, err := uc.Execute(context.Background(), dto.GetExpenseInput{ID: id})
	assert.Nil(t, out)
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrs.ErrExpenseNotFound))
}

func TestGetExpenseUC_Execute_RepoError(t *testing.T) {
	repo := mocks.NewMockExpenseRepository(t)
	uc := NewGetExpenseUC(repo)

	id := uuid.New()
	repo.EXPECT().
		Get(mock.Anything, id).
		Return(nil, errors.New("db connection failure")).
		Once()

	out, err := uc.Execute(context.Background(), dto.GetExpenseInput{ID: id})
	assert.Nil(t, out)
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrs.ErrGetExpenseDB))
}
