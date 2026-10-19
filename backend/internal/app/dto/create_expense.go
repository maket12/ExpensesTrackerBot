package dto

import "time"

type ExpenseDTO struct {
	ID        string    `json:"id"`
	UserID    int64     `json:"user_id"`
	Amount    float64   `json:"amount"`
	Category  string    `json:"category"`
	SpentAt   time.Time `json:"spent_at"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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
