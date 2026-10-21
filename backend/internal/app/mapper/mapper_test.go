package mapper

import (
	"errors"
	"math"
	"testing"
	"time"

	"backend/internal/app/dto"
	"backend/internal/domain/models"
	pkgerrs "backend/pkg/errs"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var categories = []string{"еда", "транспорт", "жильё", "развлечения", "здоровье", "покупки", "другое"}

func randomCategory() string {
	return gofakeit.RandomString(categories)
}

func createFakeDomainExpense(t *testing.T) *models.Expense {
	t.Helper()
	userID := int64(gofakeit.Number(1, 100000))
	amount := math.Round(gofakeit.Float64Range(1.0, 10000.0)*100) / 100
	category := randomCategory()
	spentAt := time.Now().UTC().Add(-time.Duration(gofakeit.Number(1, 100)) * time.Hour)
	comment := gofakeit.Sentence(3)

	e, err := models.NewExpense(userID, amount, category, spentAt, comment)
	require.NoError(t, err)
	return e
}

func TestToExpenseDTO(t *testing.T) {
	t.Run("nil expense returns empty DTO", func(t *testing.T) {
		res := ToExpenseDTO(nil)
		assert.Equal(t, dto.ExpenseDTO{}, res)
	})

	t.Run("valid expense mapped correctly", func(t *testing.T) {
		expense := createFakeDomainExpense(t)
		res := ToExpenseDTO(expense)

		assert.Equal(t, expense.ID(), res.ID)
		assert.Equal(t, expense.UserID(), res.UserID)
		assert.Equal(t, expense.Amount(), res.Amount)
		assert.Equal(t, expense.Category().String(), res.Category)
		assert.Equal(t, expense.SpentAt(), res.SpentAt)
		assert.Equal(t, expense.Comment(), res.Comment)
		assert.Equal(t, expense.CreatedAt(), res.CreatedAt)
		assert.Equal(t, expense.UpdatedAt(), res.UpdatedAt)
	})
}

func TestToExpenseDTOList(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		res := ToExpenseDTOList([]*models.Expense{})
		assert.Empty(t, res)
	})

	t.Run("populated slice with fake data", func(t *testing.T) {
		count := gofakeit.Number(2, 5)
		expenses := make([]*models.Expense, count)
		for i := 0; i < count; i++ {
			expenses[i] = createFakeDomainExpense(t)
		}

		res := ToExpenseDTOList(expenses)
		require.Len(t, res, count)
		for i := 0; i < count; i++ {
			assert.Equal(t, expenses[i].ID(), res[i].ID)
			assert.Equal(t, expenses[i].UserID(), res[i].UserID)
			assert.Equal(t, expenses[i].Amount(), res[i].Amount)
		}
	})
}

func TestToCreateExpenseOutput(t *testing.T) {
	expense := createFakeDomainExpense(t)
	out := ToCreateExpenseOutput(expense)

	require.NotNil(t, out)
	assert.Equal(t, expense.ID(), out.Expense.ID)
	assert.Equal(t, expense.UserID(), out.Expense.UserID)
	assert.Equal(t, expense.Amount(), out.Expense.Amount)
}

func TestToGetExpenseOutput(t *testing.T) {
	expense := createFakeDomainExpense(t)
	out := ToGetExpenseOutput(expense)

	require.NotNil(t, out)
	assert.Equal(t, expense.ID(), out.Expense.ID)
	assert.Equal(t, expense.Category().String(), out.Expense.Category)
}

func TestToUpdateExpenseOutput(t *testing.T) {
	expense := createFakeDomainExpense(t)
	out := ToUpdateExpenseOutput(expense)

	require.NotNil(t, out)
	assert.Equal(t, expense.ID(), out.Expense.ID)
	assert.Equal(t, expense.Comment(), out.Expense.Comment)
}

func TestToDeleteExpenseOutput(t *testing.T) {
	out := ToDeleteExpenseOutput()

	require.NotNil(t, out)
	assert.True(t, out.Success)
}

func TestToListExpensesOutput(t *testing.T) {
	count := gofakeit.Number(1, 4)
	expenses := make([]*models.Expense, count)
	for i := 0; i < count; i++ {
		expenses[i] = createFakeDomainExpense(t)
	}

	out := ToListExpensesOutput(expenses)
	require.NotNil(t, out)
	require.Len(t, out.Expenses, count)
}

func TestToGetDistinctUsersOutput(t *testing.T) {
	count := gofakeit.Number(1, 5)
	userIDs := make([]int64, count)
	for i := 0; i < count; i++ {
		userIDs[i] = int64(gofakeit.Number(100, 9999))
	}

	out := ToGetDistinctUsersOutput(userIDs)
	require.NotNil(t, out)
	assert.Equal(t, userIDs, out.UserIDs)
}

func TestToHandleAgentMessageOutput(t *testing.T) {
	reply := gofakeit.Sentence(5)
	out := ToHandleAgentMessageOutput(reply)

	require.NotNil(t, out)
	assert.Equal(t, reply, out.ResponseText)
}

func TestToDomainFilter(t *testing.T) {
	userID := int64(gofakeit.Number(1, 10000))
	cat := randomCategory()
	now := time.Now().UTC()
	from := now.Add(-24 * time.Hour)
	to := now

	t.Run("valid filter with all parameters", func(t *testing.T) {
		limit := gofakeit.Number(1, 100)
		offset := gofakeit.Number(0, 50)
		input := dto.ListExpensesInput{
			UserID:   &userID,
			Category: &cat,
			From:     &from,
			To:       &to,
			Limit:    limit,
			Offset:   offset,
		}

		filter, err := ToDomainFilter(input)
		require.NoError(t, err)
		require.NotNil(t, filter.UserID)
		assert.Equal(t, userID, *filter.UserID)
		require.NotNil(t, filter.Category)
		assert.Equal(t, cat, filter.Category.String())
		assert.Equal(t, &from, filter.From)
		assert.Equal(t, &to, filter.To)
		assert.Equal(t, limit, filter.Limit)
		assert.Equal(t, offset, filter.Offset)
	})

	t.Run("default limit and offset", func(t *testing.T) {
		input := dto.ListExpensesInput{
			Limit:  0,
			Offset: -1,
		}

		filter, err := ToDomainFilter(input)
		require.NoError(t, err)
		assert.Equal(t, 50, filter.Limit)
		assert.Equal(t, 0, filter.Offset)
	})

	t.Run("category 'все' resolves to nil", func(t *testing.T) {
		allCat := "все"
		input := dto.ListExpensesInput{
			Category: &allCat,
		}

		filter, err := ToDomainFilter(input)
		require.NoError(t, err)
		assert.Nil(t, filter.Category)
	})

	t.Run("empty category string resolves to nil", func(t *testing.T) {
		emptyCat := ""
		input := dto.ListExpensesInput{
			Category: &emptyCat,
		}

		filter, err := ToDomainFilter(input)
		require.NoError(t, err)
		assert.Nil(t, filter.Category)
	})

	t.Run("invalid category returns ValueInvalidError", func(t *testing.T) {
		invalidCat := gofakeit.Word() + "_invalid"
		input := dto.ListExpensesInput{
			Category: &invalidCat,
		}

		_, err := ToDomainFilter(input)
		require.Error(t, err)
		var valErr *pkgerrs.ValueInvalidError
		assert.True(t, errors.As(err, &valErr))
		assert.Equal(t, "category", valErr.ParamName)
	})
}
