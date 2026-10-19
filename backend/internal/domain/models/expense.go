package models

import (
	"strings"
	"time"

	"backend/internal/domain/errs"

	"github.com/google/uuid"
)

type Expense struct {
	ID        uuid.UUID
	UserID    int64
	Amount    Amount
	Category  Category
	SpentAt   time.Time
	Comment   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewExpense(
	id uuid.UUID,
	userID int64,
	amount Amount,
	category Category,
	spentAt time.Time,
	comment string,
) (*Expense, error) {
	if userID <= 0 {
		return nil, errs.ErrInvalidUserID
	}
	if !category.IsValid() {
		return nil, errs.ErrInvalidCategory
	}
	if spentAt.IsZero() {
		return nil, errs.ErrInvalidSpentAt
	}
	if id == uuid.Nil {
		id = uuid.New()
	}

	now := time.Now().UTC()
	return &Expense{
		ID:        id,
		UserID:    userID,
		Amount:    amount,
		Category:  category,
		SpentAt:   spentAt.UTC(),
		Comment:   strings.TrimSpace(comment),
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (e *Expense) Update(amount Amount, category Category, spentAt time.Time, comment string) error {
	if !category.IsValid() {
		return errs.ErrInvalidCategory
	}
	if spentAt.IsZero() {
		return errs.ErrInvalidSpentAt
	}

	e.Amount = amount
	e.Category = category
	e.SpentAt = spentAt.UTC()
	e.Comment = strings.TrimSpace(comment)
	e.UpdatedAt = time.Now().UTC()
	return nil
}
