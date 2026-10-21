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

func TestUpdateExpenseUC_Execute_Success(t *testing.T) {
	repo := mocks.NewMockExpenseRepository(t)
	uc := NewUpdateExpenseUC(repo)

	e, err := models.NewExpense(100, 300, "еда", time.Now().UTC().Add(-2*time.Hour), "пицца")
	require.NoError(t, err)

	newAmount := 450.0
	newCat := "развлечения"
	newSpentAt := time.Now().UTC().Add(-time.Hour)
	newComment := "кино"

	repo.EXPECT().
		Get(mock.Anything, e.ID()).
		Return(e, nil).
		Once()

	repo.EXPECT().
		Update(mock.Anything, mock.MatchedBy(func(updated *models.Expense) bool {
			return updated.ID() == e.ID() &&
				updated.Amount() == 450.0 &&
				updated.Category().String() == "развлечения" &&
				updated.Comment() == "кино"
		})).
		RunAndReturn(func(_ context.Context, updated *models.Expense) (*models.Expense, error) {
			return updated, nil
		}).
		Once()

	out, err := uc.Execute(context.Background(), dto.UpdateExpenseInput{
		ID:            e.ID(),
		UserID:        100,
		Amount:        &newAmount,
		Category:      &newCat,
		SpentAt:       &newSpentAt,
		Comment:       &newComment,
		EnforceUserID: true,
	})

	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, 450.0, out.Expense.Amount)
	assert.Equal(t, "развлечения", out.Expense.Category)
	assert.Equal(t, "кино", out.Expense.Comment)
}

func TestUpdateExpenseUC_Execute_InvalidID(t *testing.T) {
	repo := mocks.NewMockExpenseRepository(t)
	uc := NewUpdateExpenseUC(repo)

	out, err := uc.Execute(context.Background(), dto.UpdateExpenseInput{ID: uuid.Nil})
	assert.Nil(t, out)
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrs.ErrInvalidInput))
}

func TestUpdateExpenseUC_Execute_GetNotFound(t *testing.T) {
	repo := mocks.NewMockExpenseRepository(t)
	uc := NewUpdateExpenseUC(repo)

	id := uuid.New()
	repo.EXPECT().
		Get(mock.Anything, id).
		Return(nil, pkgerrs.NewObjectNotFoundError("expense", id)).
		Once()

	out, err := uc.Execute(context.Background(), dto.UpdateExpenseInput{ID: id})
	assert.Nil(t, out)
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrs.ErrExpenseNotFound))
}

func TestUpdateExpenseUC_Execute_GetRepoError(t *testing.T) {
	repo := mocks.NewMockExpenseRepository(t)
	uc := NewUpdateExpenseUC(repo)

	id := uuid.New()
	repo.EXPECT().
		Get(mock.Anything, id).
		Return(nil, errors.New("read failure")).
		Once()

	out, err := uc.Execute(context.Background(), dto.UpdateExpenseInput{ID: id})
	assert.Nil(t, out)
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrs.ErrGetExpenseDB))
}

func TestUpdateExpenseUC_Execute_EnforceUserID_Forbidden(t *testing.T) {
	repo := mocks.NewMockExpenseRepository(t)
	uc := NewUpdateExpenseUC(repo)

	e, err := models.NewExpense(100, 300, "еда", time.Now().UTC().Add(-time.Hour), "обед")
	require.NoError(t, err)

	repo.EXPECT().
		Get(mock.Anything, e.ID()).
		Return(e, nil).
		Once()

	newAmount := 500.0
	out, err := uc.Execute(context.Background(), dto.UpdateExpenseInput{
		ID:            e.ID(),
		UserID:        200,
		Amount:        &newAmount,
		EnforceUserID: true,
	})

	assert.Nil(t, out)
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrs.ErrForbiddenAccess))
}

func TestUpdateExpenseUC_Execute_DomainValidationError(t *testing.T) {
	tests := []struct {
		name          string
		modifyInput   func() dto.UpdateExpenseInput
		expectedError error
	}{
		{
			name: "invalid amount 0",
			modifyInput: func() dto.UpdateExpenseInput {
				zero := 0.0
				return dto.UpdateExpenseInput{Amount: &zero}
			},
			expectedError: apperrs.ErrInvalidInput,
		},
		{
			name: "invalid category",
			modifyInput: func() dto.UpdateExpenseInput {
				badCat := "неизвестная_категория"
				return dto.UpdateExpenseInput{Category: &badCat}
			},
			expectedError: apperrs.ErrInvalidInput,
		},
		{
			name: "future spent_at",
			modifyInput: func() dto.UpdateExpenseInput {
				future := time.Now().UTC().Add(time.Hour)
				return dto.UpdateExpenseInput{SpentAt: &future}
			},
			expectedError: models.ErrSpentAtAfterCreatedAt,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockExpenseRepository(t)
			uc := NewUpdateExpenseUC(repo)

			e, err := models.NewExpense(100, 300, "еда", time.Now().UTC().Add(-time.Hour), "обед")
			require.NoError(t, err)

			repo.EXPECT().
				Get(mock.Anything, e.ID()).
				Return(e, nil).
				Once()

			in := tt.modifyInput()
			in.ID = e.ID()

			out, err := uc.Execute(context.Background(), in)
			assert.Nil(t, out)
			require.Error(t, err)
			assert.True(t, errors.Is(err, tt.expectedError))
		})
	}
}

func TestUpdateExpenseUC_Execute_UpdateNotFound(t *testing.T) {
	repo := mocks.NewMockExpenseRepository(t)
	uc := NewUpdateExpenseUC(repo)

	e, err := models.NewExpense(100, 300, "еда", time.Now().UTC().Add(-time.Hour), "обед")
	require.NoError(t, err)

	repo.EXPECT().
		Get(mock.Anything, e.ID()).
		Return(e, nil).
		Once()

	repo.EXPECT().
		Update(mock.Anything, mock.Anything).
		Return(nil, pkgerrs.NewObjectNotFoundError("expense", e.ID())).
		Once()

	newAmount := 500.0
	out, err := uc.Execute(context.Background(), dto.UpdateExpenseInput{
		ID:     e.ID(),
		Amount: &newAmount,
	})

	assert.Nil(t, out)
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrs.ErrExpenseNotFound))
}

func TestUpdateExpenseUC_Execute_UpdateRepoError(t *testing.T) {
	repo := mocks.NewMockExpenseRepository(t)
	uc := NewUpdateExpenseUC(repo)

	e, err := models.NewExpense(100, 300, "еда", time.Now().UTC().Add(-time.Hour), "обед")
	require.NoError(t, err)

	repo.EXPECT().
		Get(mock.Anything, e.ID()).
		Return(e, nil).
		Once()

	repo.EXPECT().
		Update(mock.Anything, mock.Anything).
		Return(nil, errors.New("db write failed")).
		Once()

	newAmount := 500.0
	out, err := uc.Execute(context.Background(), dto.UpdateExpenseInput{
		ID:     e.ID(),
		Amount: &newAmount,
	})

	assert.Nil(t, out)
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrs.ErrUpdateExpenseDB))
}
