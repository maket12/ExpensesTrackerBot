package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"backend/internal/app/dto"
	apperrs "backend/internal/app/errs"
	"backend/internal/domain/models"
	"backend/internal/domain/port"
	"backend/internal/domain/port/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListExpensesUC_Execute_Success(t *testing.T) {
	repo := mocks.NewMockExpenseRepository(t)
	uc := NewListExpensesUC(repo)

	e1, err := models.NewExpense(100, 150, "еда", time.Now().UTC().Add(-time.Hour), "кофе")
	require.NoError(t, err)
	e2, err := models.NewExpense(100, 300, "транспорт", time.Now().UTC().Add(-30*time.Minute), "такси")
	require.NoError(t, err)

	userID := int64(100)
	cat := "еда"
	repo.EXPECT().
		List(mock.Anything, mock.MatchedBy(func(f port.ExpenseFilter) bool {
			return f.UserID != nil && *f.UserID == 100 && f.Category != nil && string(*f.Category) == "еда"
		})).
		Return([]*models.Expense{e1, e2}, nil).
		Once()

	out, err := uc.Execute(context.Background(), dto.ListExpensesInput{
		UserID:   &userID,
		Category: &cat,
	})

	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Len(t, out.Expenses, 2)
	assert.Equal(t, e1.ID(), out.Expenses[0].ID)
	assert.Equal(t, e2.ID(), out.Expenses[1].ID)
}

func TestListExpensesUC_Execute_InvalidCategory(t *testing.T) {
	repo := mocks.NewMockExpenseRepository(t)
	uc := NewListExpensesUC(repo)

	invalidCat := "невалидная_категория"
	out, err := uc.Execute(context.Background(), dto.ListExpensesInput{
		Category: &invalidCat,
	})

	assert.Nil(t, out)
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrs.ErrInvalidInput))
}

func TestListExpensesUC_Execute_RepoError(t *testing.T) {
	repo := mocks.NewMockExpenseRepository(t)
	uc := NewListExpensesUC(repo)

	repo.EXPECT().
		List(mock.Anything, mock.Anything).
		Return(nil, errors.New("db error")).
		Once()

	out, err := uc.Execute(context.Background(), dto.ListExpensesInput{})
	assert.Nil(t, out)
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrs.ErrListExpensesDB))
}
