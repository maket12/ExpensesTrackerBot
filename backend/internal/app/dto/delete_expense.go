package dto

type DeleteExpenseInput struct {
	ID            string
	UserID        int64
	EnforceUserID bool
}

type DeleteExpenseOutput struct {
	Success bool
}
