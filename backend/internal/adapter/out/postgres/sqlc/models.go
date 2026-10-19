package sqlc

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Expense struct {
	ID        pgtype.UUID
	UserID    int64
	Amount    pgtype.Numeric
	Category  string
	SpentAt   pgtype.Timestamptz
	Comment   string
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
}
