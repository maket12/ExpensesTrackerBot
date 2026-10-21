package dto

import "github.com/google/uuid"

type GetExpenseInput struct {
	ID uuid.UUID
}

type GetExpenseOutput struct {
	Expense ExpenseDTO
}
