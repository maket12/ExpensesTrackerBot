package dto

import (
	"time"

	"github.com/google/uuid"
)

type UpdateExpenseInput struct {
	ID            uuid.UUID
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
