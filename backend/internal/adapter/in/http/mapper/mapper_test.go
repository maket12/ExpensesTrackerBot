package mapper

import (
	"errors"
	"math"
	"testing"
	"time"

	httpdto "backend/internal/adapter/in/http/dto"
	appdto "backend/internal/app/dto"
	pkgerrs "backend/pkg/errs"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var categories = []string{"еда", "транспорт", "жильё", "развлечения", "здоровье", "покупки", "другое"}

func randomCategory() string {
	return gofakeit.RandomString(categories)
}

func TestToCreateExpenseInput(t *testing.T) {
	req := httpdto.CreateExpenseRequest{
		UserID:   int64(gofakeit.Number(1, 10000)),
		Amount:   math.Round(gofakeit.Float64Range(1.0, 1000.0)*100) / 100,
		Category: randomCategory(),
		SpentAt:  time.Now().UTC().Add(-time.Hour),
		Comment:  gofakeit.Sentence(2),
	}

	input := ToCreateExpenseInput(req)
	assert.Equal(t, req.UserID, input.UserID)
	assert.Equal(t, req.Amount, input.Amount)
	assert.Equal(t, req.Category, input.Category)
	assert.Equal(t, req.SpentAt, input.SpentAt)
	assert.Equal(t, req.Comment, input.Comment)
}

func TestToGetExpenseInput(t *testing.T) {
	t.Run("valid uuid string", func(t *testing.T) {
		expectedID := uuid.New()
		input, err := ToGetExpenseInput(expectedID.String())

		require.NoError(t, err)
		assert.Equal(t, expectedID, input.ID)
	})

	t.Run("invalid uuid string", func(t *testing.T) {
		invalidID := gofakeit.Word() + "-invalid"
		_, err := ToGetExpenseInput(invalidID)

		require.Error(t, err)
		var valErr *pkgerrs.ValueInvalidError
		assert.True(t, errors.As(err, &valErr))
		assert.Equal(t, "id", valErr.ParamName)
	})
}

func TestToUpdateExpenseInput(t *testing.T) {
	t.Run("valid uuid string with update fields", func(t *testing.T) {
		expectedID := uuid.New()
		amount := math.Round(gofakeit.Float64Range(10.0, 500.0)*100) / 100
		cat := randomCategory()
		spentAt := time.Now().UTC().Add(-2 * time.Hour)
		comm := gofakeit.Sentence(2)

		req := httpdto.UpdateExpenseRequest{
			UserID:   int64(gofakeit.Number(1, 10000)),
			Amount:   &amount,
			Category: &cat,
			SpentAt:  &spentAt,
			Comment:  &comm,
		}

		input, err := ToUpdateExpenseInput(expectedID.String(), req)
		require.NoError(t, err)
		assert.Equal(t, expectedID, input.ID)
		assert.Equal(t, req.UserID, input.UserID)
		assert.Equal(t, req.Amount, input.Amount)
		assert.Equal(t, req.Category, input.Category)
		assert.Equal(t, req.SpentAt, input.SpentAt)
		assert.Equal(t, req.Comment, input.Comment)
		assert.False(t, input.EnforceUserID)
	})

	t.Run("invalid uuid string", func(t *testing.T) {
		invalidID := gofakeit.Word()
		_, err := ToUpdateExpenseInput(invalidID, httpdto.UpdateExpenseRequest{})

		require.Error(t, err)
		var valErr *pkgerrs.ValueInvalidError
		assert.True(t, errors.As(err, &valErr))
		assert.Equal(t, "id", valErr.ParamName)
	})
}

func TestToDeleteExpenseInput(t *testing.T) {
	t.Run("valid uuid string", func(t *testing.T) {
		expectedID := uuid.New()
		input, err := ToDeleteExpenseInput(expectedID.String())

		require.NoError(t, err)
		assert.Equal(t, expectedID, input.ID)
		assert.False(t, input.EnforceUserID)
		assert.Equal(t, int64(0), input.UserID)
	})

	t.Run("invalid uuid string", func(t *testing.T) {
		invalidID := gofakeit.Word()
		_, err := ToDeleteExpenseInput(invalidID)

		require.Error(t, err)
		var valErr *pkgerrs.ValueInvalidError
		assert.True(t, errors.As(err, &valErr))
		assert.Equal(t, "id", valErr.ParamName)
	})
}

func TestToExpenseResponse(t *testing.T) {
	d := appdto.ExpenseDTO{
		ID:        uuid.New(),
		UserID:    int64(gofakeit.Number(1, 10000)),
		Amount:    math.Round(gofakeit.Float64Range(1.0, 5000.0)*100) / 100,
		Category:  randomCategory(),
		SpentAt:   time.Now().UTC().Add(-time.Hour),
		Comment:   gofakeit.Sentence(2),
		CreatedAt: time.Now().UTC().Add(-time.Minute),
		UpdatedAt: time.Now().UTC(),
	}

	resp := ToExpenseResponse(d)
	assert.Equal(t, d.ID.String(), resp.ID)
	assert.Equal(t, d.UserID, resp.UserID)
	assert.Equal(t, d.Amount, resp.Amount)
	assert.Equal(t, d.Category, resp.Category)
	assert.Equal(t, d.SpentAt, resp.SpentAt)
	assert.Equal(t, d.Comment, resp.Comment)
	assert.Equal(t, d.CreatedAt, resp.CreatedAt)
	assert.Equal(t, d.UpdatedAt, resp.UpdatedAt)
}

func TestToExpenseResponses(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		resp := ToExpenseResponses([]appdto.ExpenseDTO{})
		assert.Empty(t, resp)
	})

	t.Run("populated slice", func(t *testing.T) {
		count := gofakeit.Number(2, 5)
		dtos := make([]appdto.ExpenseDTO, count)
		for i := 0; i < count; i++ {
			dtos[i] = appdto.ExpenseDTO{
				ID:       uuid.New(),
				UserID:   int64(gofakeit.Number(1, 10000)),
				Amount:   math.Round(gofakeit.Float64Range(1.0, 100.0)*100) / 100,
				Category: randomCategory(),
			}
		}

		resps := ToExpenseResponses(dtos)
		require.Len(t, resps, count)
		for i := 0; i < count; i++ {
			assert.Equal(t, dtos[i].ID.String(), resps[i].ID)
			assert.Equal(t, dtos[i].Amount, resps[i].Amount)
		}
	})
}
