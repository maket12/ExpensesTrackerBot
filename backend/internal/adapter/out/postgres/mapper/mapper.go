package mapper

import (
	"fmt"
	"time"

	"backend/internal/adapter/out/postgres/sqlc"
	"backend/internal/domain/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func ToPgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: id != uuid.Nil}
}

func FromPgUUID(p pgtype.UUID) uuid.UUID {
	return uuid.UUID(p.Bytes)
}

func ToPgNumeric(val float64) pgtype.Numeric {
	var num pgtype.Numeric
	_ = num.Scan(fmt.Sprintf("%.2f", val))
	return num
}

func FromPgNumeric(num pgtype.Numeric) float64 {
	f, _ := num.Float64Value()
	return f.Float64
}

func ToPgTimestamptz(t time.Time) pgtype.Timestamptz {
	if t.IsZero() {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: t.UTC(), Valid: true}
}

func FromPgTimestamptz(tz pgtype.Timestamptz) time.Time {
	return tz.Time
}

func ToDomainExpense(row sqlc.Expense) (*models.Expense, error) {
	amount, err := models.NewAmount(FromPgNumeric(row.Amount))
	if err != nil {
		return nil, err
	}

	category, err := models.NewCategory(row.Category)
	if err != nil {
		return nil, err
	}

	return &models.Expense{
		ID:        FromPgUUID(row.ID),
		UserID:    row.UserID,
		Amount:    amount,
		Category:  category,
		SpentAt:   FromPgTimestamptz(row.SpentAt),
		Comment:   row.Comment,
		CreatedAt: FromPgTimestamptz(row.CreatedAt),
		UpdatedAt: FromPgTimestamptz(row.UpdatedAt),
	}, nil
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
		ID:        ToPgUUID(e.ID),
		UserID:    e.UserID,
		Amount:    ToPgNumeric(e.Amount.Value()),
		Category:  e.Category.String(),
		SpentAt:   ToPgTimestamptz(e.SpentAt),
		Comment:   e.Comment,
		CreatedAt: ToPgTimestamptz(e.CreatedAt),
		UpdatedAt: ToPgTimestamptz(e.UpdatedAt),
	}
}

func ToUpdateExpenseParams(e *models.Expense) sqlc.UpdateExpenseParams {
	return sqlc.UpdateExpenseParams{
		ID:        ToPgUUID(e.ID),
		Amount:    ToPgNumeric(e.Amount.Value()),
		Category:  e.Category.String(),
		SpentAt:   ToPgTimestamptz(e.SpentAt),
		Comment:   e.Comment,
		UpdatedAt: ToPgTimestamptz(e.UpdatedAt),
	}
}

func ToListExpensesFilteredParams(filter models.ExpenseFilter) sqlc.ListExpensesFilteredParams {
	var categoryStr *string
	if filter.Category != nil {
		s := filter.Category.String()
		categoryStr = &s
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
