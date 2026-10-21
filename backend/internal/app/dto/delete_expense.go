package dto

import "github.com/google/uuid"

type DeleteExpenseInput struct {
	ID            uuid.UUID
	UserID        int64
	EnforceUserID bool
}

type DeleteExpenseOutput struct {
	Success bool
}
