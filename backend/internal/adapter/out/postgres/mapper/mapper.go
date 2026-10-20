package mapper

import (
	"time"

	"backend/internal/adapter/out/postgres/sqlc"
	"backend/internal/domain/models"
	pkgpostgres "backend/pkg/postgres"
	"backend/pkg/utils"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func ToPgUUID(id uuid.UUID) pgtype.UUID {
	return pkgpostgres.MapToPgUUID(id)
}

func FromPgUUID(p pgtype.UUID) uuid.UUID {
	return uuid.UUID(p.Bytes)
}

func ToPgNumeric(val float64) pgtype.Numeric {
	n, _ := pkgpostgres.MapFloat64ToPgNumeric(val, nil)
	return n
}

func FromPgNumeric(num pgtype.Numeric) (float64, error) {
	return pkgpostgres.MapPgNumericToFloat64(num)
}

func ToPgTimestamptz(t time.Time) pgtype.Timestamptz {
	if t.IsZero() {
		return pgtype.Timestamptz{Valid: false}
	}
	return pkgpostgres.MapToPgTimestamptz(t.UTC())
}

func FromPgTimestamptz(tz pgtype.Timestamptz) time.Time {
	return tz.Time
}

func ToDomainExpense(row sqlc.Expense) (*models.Expense, error) {
	amountVal, err := FromPgNumeric(row.Amount)
	if err != nil {
		return nil, err
	}

	amount, err := models.NewAmount(amountVal)
	if err != nil {
		return nil, err
	}

	category, err := models.NewCategory(row.Category)
	if err != nil {
		return nil, err
	}

	return models.ReconstituteExpense(
		FromPgUUID(row.ID),
		row.UserID,
		amount,
		category,
		FromPgTimestamptz(row.SpentAt),
		row.Comment,
		FromPgTimestamptz(row.CreatedAt),
		FromPgTimestamptz(row.UpdatedAt),
	), nil
}

func ToDomainExpenses(rows []sqlc.Expense) ([]*models.Expense, error) {
	result := make([]*models.Expense, 0, len(rows))
	for _, row := range rows {
		e, err := ToDomainExpense(row)
		if err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, nil
}

func ToCreateExpenseParams(e *models.Expense) sqlc.CreateExpenseParams {
	return sqlc.CreateExpenseParams{
		ID:        ToPgUUID(e.ID()),
		UserID:    e.UserID(),
		Amount:    ToPgNumeric(e.Amount().Value()),
		Category:  e.Category().String(),
		SpentAt:   ToPgTimestamptz(e.SpentAt()),
		Comment:   e.Comment(),
		CreatedAt: ToPgTimestamptz(e.CreatedAt()),
		UpdatedAt: ToPgTimestamptz(e.UpdatedAt()),
	}
}

func ToUpdateExpenseParams(e *models.Expense) sqlc.UpdateExpenseParams {
	return sqlc.UpdateExpenseParams{
		ID:        ToPgUUID(e.ID()),
		Amount:    ToPgNumeric(e.Amount().Value()),
		Category:  e.Category().String(),
		SpentAt:   ToPgTimestamptz(e.SpentAt()),
		Comment:   e.Comment(),
		UpdatedAt: ToPgTimestamptz(e.UpdatedAt()),
	}
}

func ToListExpensesFilteredParams(filter models.ExpenseFilter) sqlc.ListExpensesFilteredParams {
	var categoryStr *string
	if filter.Category != nil {
		categoryStr = utils.VPtr(filter.Category.String())
	}

	var fromTz, toTz pgtype.Timestamptz
	if filter.From != nil {
		fromTz = ToPgTimestamptz(*filter.From)
	}
	if filter.To != nil {
		toTz = ToPgTimestamptz(*filter.To)
	}

	limit := int32(filter.Limit)
	if limit <= 0 {
		limit = 50
	}
	offset := int32(filter.Offset)
	if offset < 0 {
		offset = 0
	}

	return sqlc.ListExpensesFilteredParams{
		UserID:   filter.UserID,
		Category: categoryStr,
		From:     fromTz,
		To:       toTz,
		Limit:    limit,
		Offset:   offset,
	}
}
