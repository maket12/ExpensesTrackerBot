package dto

import "time"

type CreateExpenseRequest struct {
	UserID   int64     `json:"user_id" example:"100"`
	Amount   float64   `json:"amount" example:"350.50"`
	Category string    `json:"category" example:"еда" enums:"еда,транспорт,жильё,развлечения,здоровье,покупки,другое"`
	SpentAt  time.Time `json:"spent_at" format:"date-time" example:"2026-10-22T15:30:00Z"`
	Comment  string    `json:"comment" example:"обед"`
}

type CreateExpenseResponse struct {
	Expense Expense `json:"expense"`
}

type GetExpenseRequest struct {
	ID string `param:"id" json:"-"`
}

type GetExpenseResponse struct {
	Expense Expense `json:"expense"`
}

type UpdateExpenseRequest struct {
	ID       string     `param:"id" json:"-"`
	UserID   int64      `json:"user_id" example:"100"`
	Amount   *float64   `json:"amount,omitempty" example:"450.00"`
	Category *string    `json:"category,omitempty" example:"еда" enums:"еда,транспорт,жильё,развлечения,здоровье,покупки,другое"`
	SpentAt  *time.Time `json:"spent_at,omitempty" format:"date-time" example:"2026-10-22T15:30:00Z"`
	Comment  *string    `json:"comment,omitempty" example:"ужин"`
}

type UpdateExpenseResponse struct {
	Expense Expense `json:"expense"`
}

type DeleteExpenseRequest struct {
	ID string `param:"id"`
}

type DeleteExpenseResponse struct{}

type ListExpensesRequest struct {
	UserID   *int64  `query:"user_id"`
	Category *string `query:"category"`
	From     *string `query:"from"`
	To       *string `query:"to"`
	Limit    *int    `query:"limit"`
	Offset   *int    `query:"offset"`
}

type ListExpensesResponse struct {
	Items []Expense `json:"items"`
}

type GetDistinctUsersResponse struct {
	UserIDs []int64 `json:"user_ids"`
}

type Expense struct {
	ID        string    `json:"id"`
	UserID    int64     `json:"user_id"`
	Amount    float64   `json:"amount"`
	Category  string    `json:"category"`
	SpentAt   time.Time `json:"spent_at"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
