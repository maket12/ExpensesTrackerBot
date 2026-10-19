package dto

import "time"

type ExpenseResponse struct {
	ID        string
	UserID    int64
	Amount    float64
	Category  string
	SpentAt   time.Time
	Comment   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SaveExpenseRequest struct {
	ID        string
	UserID    int64
	Amount    float64
	Category  string
	SpentAt   time.Time
	Comment   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type FilterExpensesRequest struct {
	UserID   *int64
	Category *string
	From     *time.Time
	To       *time.Time
	Limit    int
	Offset   int
}
