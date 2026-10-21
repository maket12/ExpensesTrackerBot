package dto

import "time"

type CreateExpenseRequest struct {
	UserID   int64     `json:"user_id"`
	Amount   float64   `json:"amount"`
	Category string    `json:"category"`
	SpentAt  time.Time `json:"spent_at"`
	Comment  string    `json:"comment"`
}

type UpdateExpenseRequest struct {
	UserID   int64      `json:"user_id"`
	Amount   *float64   `json:"amount,omitempty"`
	Category *string    `json:"category,omitempty"`
	SpentAt  *time.Time `json:"spent_at,omitempty"`
	Comment  *string    `json:"comment,omitempty"`
}

type ExpenseResponse struct {
	ID        string    `json:"id"`
	UserID    int64     `json:"user_id"`
	Amount    float64   `json:"amount"`
	Category  string    `json:"category"`
	SpentAt   time.Time `json:"spent_at"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListExpensesResponse struct {
	Items []ExpenseResponse `json:"items"`
}

type GetDistinctUsersResponse struct {
	UserIDs []int64 `json:"user_ids"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
