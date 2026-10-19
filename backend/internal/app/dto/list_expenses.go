package dto

import "time"

type ListExpensesInput struct {
	UserID   *int64
	Category *string
	From     *time.Time
	To       *time.Time
	Limit    int
	Offset   int
}

type ListExpensesOutput struct {
	Expenses []ExpenseDTO
}
