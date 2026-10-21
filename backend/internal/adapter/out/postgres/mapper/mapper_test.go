package mapper

import (
	"math"
	"testing"
	"time"

	"backend/internal/adapter/out/postgres/sqlc"
	"backend/internal/domain/models"
	"backend/internal/domain/port"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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
	spentAt := time.Now().UTC().Add(-time.Duration(gofakeit.Number(1, 100)) * time.Hour).Truncate(time.Microsecond)
	comment := gofakeit.Sentence(3)

	e, err := models.NewExpense(userID, amount, category, spentAt, comment)
	require.NoError(t, err)
	return e
}

func TestMapperUUID(t *testing.T) {
	id := uuid.MustParse(gofakeit.UUID())
	pgUUID := ToPgUUID(id)
	assert.True(t, pgUUID.Valid)

	converted := FromPgUUID(pgUUID)
	assert.Equal(t, id, converted)
}

func TestMapperNumeric(t *testing.T) {
	val := math.Round(gofakeit.Float64Range(1.0, 50000.0)*100) / 100
	pgNum := ToPgNumeric(val)
	assert.True(t, pgNum.Valid)

	got, err := FromPgNumeric(pgNum)
	require.NoError(t, err)
	assert.Equal(t, val, got)
}

func TestMapperTimestamptz(t *testing.T) {
	now := gofakeit.Date().UTC().Truncate(time.Microsecond)
	pgTime := ToPgTimestamptz(now)
	assert.True(t, pgTime.Valid)

	got := FromPgTimestamptz(pgTime)
	assert.True(t, got.Equal(now))

	zeroTime := ToPgTimestamptz(time.Time{})
	assert.False(t, zeroTime.Valid)
}

func TestToCreateExpenseParams(t *testing.T) {
	expense := createFakeDomainExpense(t)
	params := ToCreateExpenseParams(expense)

	assert.Equal(t, expense.ID(), FromPgUUID(params.ID))
	assert.Equal(t, expense.UserID(), params.UserID)
	val, err := FromPgNumeric(params.Amount)
	require.NoError(t, err)
	assert.Equal(t, expense.Amount(), val)
	assert.Equal(t, expense.Category().String(), params.Category)
	assert.Equal(t, expense.Comment(), params.Comment)
	assert.True(t, expense.SpentAt().Equal(FromPgTimestamptz(params.SpentAt)))
}

func TestToUpdateExpenseParams(t *testing.T) {
	expense := createFakeDomainExpense(t)
	params := ToUpdateExpenseParams(expense)

	assert.Equal(t, expense.ID(), FromPgUUID(params.ID))
	val, err := FromPgNumeric(params.Amount)
	require.NoError(t, err)
	assert.Equal(t, expense.Amount(), val)
	assert.Equal(t, expense.Category().String(), params.Category)
	assert.Equal(t, expense.Comment(), params.Comment)
}

func TestToListLatestExpensesParams(t *testing.T) {
	userID := int64(gofakeit.Number(1, 10000))
	limit := gofakeit.Number(1, 50)

	params := ToListLatestExpensesParams(userID, limit)
	assert.Equal(t, userID, params.UserID)
	assert.Equal(t, int32(limit), params.Limit)

	defaultParams := ToListLatestExpensesParams(userID, 0)
	assert.Equal(t, int32(5), defaultParams.Limit)
}

func TestToListExpensesFilteredParams(t *testing.T) {
	userID := int64(gofakeit.Number(1, 10000))
	cat := models.Category(randomCategory())
	from := gofakeit.Date().UTC().Add(-48 * time.Hour)
	to := gofakeit.Date().UTC()
	limit := gofakeit.Number(1, 100)
	offset := gofakeit.Number(0, 50)

	t.Run("populated filter", func(t *testing.T) {
		filter := port.ExpenseFilter{
			UserID:   &userID,
			Category: &cat,
			From:     &from,
			To:       &to,
			Limit:    limit,
			Offset:   offset,
		}

		params := ToListExpensesFilteredParams(filter)
		require.NotNil(t, params.UserID)
		assert.Equal(t, userID, *params.UserID)
		require.NotNil(t, params.Category)
		assert.Equal(t, cat.String(), *params.Category)
		assert.True(t, params.From.Valid)
		assert.True(t, params.To.Valid)
		assert.Equal(t, int32(limit), params.Limit)
		assert.Equal(t, int32(offset), params.Offset)
	})

	t.Run("empty filter with defaults", func(t *testing.T) {
		filter := port.ExpenseFilter{
			Limit:  0,
			Offset: -1,
		}

		params := ToListExpensesFilteredParams(filter)
		assert.Nil(t, params.UserID)
		assert.Nil(t, params.Category)
		assert.False(t, params.From.Valid)
		assert.False(t, params.To.Valid)
		assert.Equal(t, int32(50), params.Limit)
		assert.Equal(t, int32(0), params.Offset)
	})
}

func TestToDomainExpense_And_ToDomainExpenses(t *testing.T) {
	expense := createFakeDomainExpense(t)
	createParams := ToCreateExpenseParams(expense)

	row := sqlc.Expense{
		ID:        createParams.ID,
		UserID:    createParams.UserID,
		Amount:    createParams.Amount,
		Category:  createParams.Category,
		SpentAt:   createParams.SpentAt,
		Comment:   createParams.Comment,
		CreatedAt: createParams.CreatedAt,
		UpdatedAt: createParams.UpdatedAt,
	}

	t.Run("single row valid mapping", func(t *testing.T) {
		domainExpense, err := ToDomainExpense(row)
		require.NoError(t, err)
		assert.Equal(t, expense.ID(), domainExpense.ID())
		assert.Equal(t, expense.UserID(), domainExpense.UserID())
		assert.Equal(t, expense.Amount(), domainExpense.Amount())
		assert.Equal(t, expense.Category(), domainExpense.Category())
		assert.Equal(t, expense.Comment(), domainExpense.Comment())
	})

	t.Run("multiple rows valid mapping", func(t *testing.T) {
		rows := []sqlc.Expense{row, row}
		list, err := ToDomainExpenses(rows)
		require.NoError(t, err)
		require.Len(t, list, 2)
	})

	t.Run("empty rows slice", func(t *testing.T) {
		list, err := ToDomainExpenses([]sqlc.Expense{})
		require.NoError(t, err)
		assert.Empty(t, list)
	})

	t.Run("invalid numeric returns error", func(t *testing.T) {
		invalidRow := row
		invalidRow.Amount = pgtype.Numeric{Valid: false}

		_, err := ToDomainExpense(invalidRow)
		require.Error(t, err)

		_, err = ToDomainExpenses([]sqlc.Expense{invalidRow})
		require.Error(t, err)
	})
}
