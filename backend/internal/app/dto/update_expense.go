package dto

import "time"

type UpdateExpenseInput struct {
	ID            string
	UserID        int64
	Amount        *float64
	Category      *string
	SpentAt       *time.Time
	Comment       *string
	EnforceUserID bool
}

type UpdateExpenseOutput struct {
	Expense ExpenseDTO
}
