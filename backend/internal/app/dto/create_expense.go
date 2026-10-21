package dto

import (
	"time"

	"github.com/google/uuid"
)

type ExpenseDTO struct {
	ID        uuid.UUID
	UserID    int64
	Amount    float64
	Category  string
	SpentAt   time.Time
	Comment   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CreateExpenseInput struct {
	UserID   int64
	Amount   float64
	Category string
	SpentAt  time.Time
	Comment  string
}

type CreateExpenseOutput struct {
	Expense ExpenseDTO
}
