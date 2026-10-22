package usecases

import (
	"context"
	"errors"
	"testing"

	apperrs "backend/internal/app/errs"
	"backend/internal/domain/port/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetDistinctUsersUC_Execute_Success(t *testing.T) {
	repo := mocks.NewMockExpenseRepository(t)
	uc := NewGetDistinctUsersUC(repo)

	expectedIDs := []int64{100, 200, 300}
	repo.EXPECT().
		ListUserIDS(mock.Anything).
		Return(expectedIDs, nil).
		Once()

	out, err := uc.Execute(context.Background())
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, expectedIDs, out.UserIDs)
}

func TestGetDistinctUsersUC_Execute_RepoError(t *testing.T) {
	repo := mocks.NewMockExpenseRepository(t)
	uc := NewGetDistinctUsersUC(repo)

	repo.EXPECT().
		ListUserIDS(mock.Anything).
		Return(nil, errors.New("db error")).
		Once()

	out, err := uc.Execute(context.Background())
	assert.Nil(t, out)
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrs.ErrListUserIDSDB))
}
