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

func TestToCreateExpenseResponse(t *testing.T) {
	d := appdto.ExpenseDTO{
		ID:        uuid.New(),
		UserID:    int64(gofakeit.Number(1, 10000)),
		Amount:    123.45,
		Category:  "еда",
		SpentAt:   time.Now().UTC(),
		Comment:   "lunch",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	resp := ToCreateExpenseResponse(d)
	assert.Equal(t, d.ID.String(), resp.Expense.ID)
	assert.Equal(t, d.UserID, resp.Expense.UserID)
	assert.Equal(t, d.Amount, resp.Expense.Amount)
}

func TestToGetExpenseInput(t *testing.T) {
	t.Run("valid uuid string", func(t *testing.T) {
		expectedID := uuid.New()
		input, err := ToGetExpenseInput(httpdto.GetExpenseRequest{ID: expectedID.String()})

		require.NoError(t, err)
		assert.Equal(t, expectedID, input.ID)
	})

	t.Run("invalid uuid string", func(t *testing.T) {
		invalidID := gofakeit.Word() + "-invalid"
		_, err := ToGetExpenseInput(httpdto.GetExpenseRequest{ID: invalidID})

		require.Error(t, err)
		var valErr *pkgerrs.ValueInvalidError
		assert.True(t, errors.As(err, &valErr))
		assert.Equal(t, "id", valErr.ParamName)
	})
}

func TestToGetExpenseResponse(t *testing.T) {
	d := appdto.ExpenseDTO{
		ID:     uuid.New(),
		UserID: 123,
		Amount: 456.78,
	}

	resp := ToGetExpenseResponse(d)
	assert.Equal(t, d.ID.String(), resp.Expense.ID)
	assert.Equal(t, d.UserID, resp.Expense.UserID)
	assert.Equal(t, d.Amount, resp.Expense.Amount)
}

func TestToUpdateExpenseInput(t *testing.T) {
	t.Run("valid uuid string with update fields", func(t *testing.T) {
		expectedID := uuid.New()
		amount := math.Round(gofakeit.Float64Range(10.0, 500.0)*100) / 100
		cat := randomCategory()
		spentAt := time.Now().UTC().Add(-2 * time.Hour)
		comm := gofakeit.Sentence(2)

		req := httpdto.UpdateExpenseRequest{
			ID:       expectedID.String(),
			UserID:   int64(gofakeit.Number(1, 10000)),
			Amount:   &amount,
			Category: &cat,
			SpentAt:  &spentAt,
			Comment:  &comm,
		}

		input, err := ToUpdateExpenseInput(req)
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
		_, err := ToUpdateExpenseInput(httpdto.UpdateExpenseRequest{ID: invalidID})

		require.Error(t, err)
		var valErr *pkgerrs.ValueInvalidError
		assert.True(t, errors.As(err, &valErr))
		assert.Equal(t, "id", valErr.ParamName)
	})
}

func TestToUpdateExpenseResponse(t *testing.T) {
	d := appdto.ExpenseDTO{
		ID:     uuid.New(),
		UserID: 123,
		Amount: 789.0,
	}

	resp := ToUpdateExpenseResponse(d)
	assert.Equal(t, d.ID.String(), resp.Expense.ID)
	assert.Equal(t, d.Amount, resp.Expense.Amount)
}

func TestToDeleteExpenseInput(t *testing.T) {
	t.Run("valid uuid string", func(t *testing.T) {
		expectedID := uuid.New()
		input, err := ToDeleteExpenseInput(httpdto.DeleteExpenseRequest{ID: expectedID.String()})

		require.NoError(t, err)
		assert.Equal(t, expectedID, input.ID)
		assert.False(t, input.EnforceUserID)
		assert.Equal(t, int64(0), input.UserID)
	})

	t.Run("invalid uuid string", func(t *testing.T) {
		invalidID := gofakeit.Word()
		_, err := ToDeleteExpenseInput(httpdto.DeleteExpenseRequest{ID: invalidID})

		require.Error(t, err)
		var valErr *pkgerrs.ValueInvalidError
		assert.True(t, errors.As(err, &valErr))
		assert.Equal(t, "id", valErr.ParamName)
	})
}

func TestToListExpensesInput(t *testing.T) {
	t.Run("valid filters", func(t *testing.T) {
		uid := int64(42)
		cat := "еда"
		fromStr := "2026-01-01T00:00:00Z"
		toStr := "2026-01-31T23:59:59Z"
		limit := 10
		offset := 20

		req := httpdto.ListExpensesRequest{
			UserID:   &uid,
			Category: &cat,
			From:     &fromStr,
			To:       &toStr,
			Limit:    &limit,
			Offset:   &offset,
		}

		input, err := ToListExpensesInput(req)
		require.NoError(t, err)
		assert.Equal(t, &uid, input.UserID)
		assert.Equal(t, &cat, input.Category)
		assert.NotNil(t, input.From)
		assert.NotNil(t, input.To)
		assert.Equal(t, 10, input.Limit)
		assert.Equal(t, 20, input.Offset)
	})

	t.Run("invalid from timestamp", func(t *testing.T) {
		fromStr := "invalid-time"
		req := httpdto.ListExpensesRequest{
			From: &fromStr,
		}

		_, err := ToListExpensesInput(req)
		require.Error(t, err)
	})

	t.Run("invalid to timestamp", func(t *testing.T) {
		toStr := "invalid-time"
		req := httpdto.ListExpensesRequest{
			To: &toStr,
		}

		_, err := ToListExpensesInput(req)
		require.Error(t, err)
	})
}

func TestToListExpensesResponse(t *testing.T) {
	dtos := []appdto.ExpenseDTO{
		{ID: uuid.New(), UserID: 1, Amount: 100},
		{ID: uuid.New(), UserID: 2, Amount: 200},
	}

	resp := ToListExpensesResponse(dtos)
	assert.Len(t, resp.Items, 2)
	assert.Len(t, resp.Expenses, 2)
	assert.Equal(t, dtos[0].ID.String(), resp.Items[0].ID)
	assert.Equal(t, dtos[1].ID.String(), resp.Items[1].ID)
}

func TestToGetDistinctUsersInputAndResponse(t *testing.T) {
	in := ToGetDistinctUsersInput(httpdto.GetDistinctUsersRequest{})
	assert.Equal(t, appdto.GetDistinctUsersInput{}, in)

	resp := ToGetDistinctUsersResponse([]int64{1, 2, 3})
	assert.Equal(t, []int64{1, 2, 3}, resp.UserIDs)

	respNil := ToGetDistinctUsersResponse(nil)
	assert.Empty(t, respNil.UserIDs)
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
