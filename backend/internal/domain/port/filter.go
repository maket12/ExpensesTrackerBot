package port

import (
	"time"

	"backend/internal/domain/models"
)

type ExpenseFilter struct {
	UserID   *int64
	Category *models.Category
	From     *time.Time
	To       *time.Time
	Limit    int
	Offset   int
}
