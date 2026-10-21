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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateExpenseUC_Execute_Success(t *testing.T) {
	repo := mocks.NewMockExpenseRepository(t)
	uc := NewCreateExpenseUC(repo)

	spentAt := time.Now().UTC().Add(-time.Hour)
	input := dto.CreateExpenseInput{
		UserID:   100,
		Amount:   250.50,
		Category: "еда",
		SpentAt:  spentAt,
		Comment:  "обед",
	}

	repo.EXPECT().
		Create(mock.Anything, mock.MatchedBy(func(e *models.Expense) bool {
			return e.UserID() == 100 && e.Amount() == 250.50 && e.Category().String() == "еда"
		})).
		RunAndReturn(func(_ context.Context, e *models.Expense) (*models.Expense, error) {
			return e, nil
		}).
		Once()

	out, err := uc.Execute(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.NotEmpty(t, out.Expense.ID)
	assert.Equal(t, int64(100), out.Expense.UserID)
	assert.Equal(t, 250.50, out.Expense.Amount)
	assert.Equal(t, "еда", out.Expense.Category)
	assert.Equal(t, "обед", out.Expense.Comment)
}

func TestCreateExpenseUC_Execute_ValidationError(t *testing.T) {
	tests := []struct {
		name  string
		input dto.CreateExpenseInput
	}{
		{
			name: "negative amount",
			input: dto.CreateExpenseInput{
				UserID:   100,
				Amount:   -10,
				Category: "еда",
				SpentAt:  time.Now().UTC().Add(-time.Minute),
			},
		},
		{
			name: "amount below min",
			input: dto.CreateExpenseInput{
				UserID:   100,
				Amount:   0.5,
				Category: "еда",
				SpentAt:  time.Now().UTC().Add(-time.Minute),
			},
		},
		{
			name: "amount above max",
			input: dto.CreateExpenseInput{
				UserID:   100,
				Amount:   2_000_000,
				Category: "еда",
				SpentAt:  time.Now().UTC().Add(-time.Minute),
			},
		},
		{
			name: "invalid user_id",
			input: dto.CreateExpenseInput{
				UserID:   0,
				Amount:   100,
				Category: "еда",
				SpentAt:  time.Now().UTC().Add(-time.Minute),
			},
		},
		{
			name: "invalid category",
			input: dto.CreateExpenseInput{
				UserID:   100,
				Amount:   100,
				Category: "неизвестно",
				SpentAt:  time.Now().UTC().Add(-time.Minute),
			},
		},
		{
			name: "future spent_at",
			input: dto.CreateExpenseInput{
				UserID:   100,
				Amount:   100,
				Category: "еда",
				SpentAt:  time.Now().UTC().Add(time.Hour),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewMockExpenseRepository(t)
			uc := NewCreateExpenseUC(repo)

			out, err := uc.Execute(context.Background(), tt.input)
			assert.Nil(t, out)
			require.Error(t, err)
			assert.True(t, errors.Is(err, apperrs.ErrInvalidInput))
		})
	}
}

func TestCreateExpenseUC_Execute_RepoError(t *testing.T) {
	repo := mocks.NewMockExpenseRepository(t)
	uc := NewCreateExpenseUC(repo)

	input := dto.CreateExpenseInput{
		UserID:   100,
		Amount:   250,
		Category: "еда",
		SpentAt:  time.Now().UTC().Add(-time.Minute),
	}

	repo.EXPECT().
		Create(mock.Anything, mock.Anything).
		Return(nil, errors.New("connection failed")).
		Once()

	out, err := uc.Execute(context.Background(), input)
	assert.Nil(t, out)
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrs.ErrCreateExpenseDB))
}
