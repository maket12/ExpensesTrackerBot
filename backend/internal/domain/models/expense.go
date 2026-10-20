package models

import (
	pkgerrs "backend/pkg/errs"
	"strings"
	"time"

	"backend/internal/domain/errs"

	"github.com/google/uuid"
)

// ================ Rich model for Expense ================

type Expense struct {
	id        uuid.UUID
	userID    int64
	amount    Amount
	category  Category
	spentAt   time.Time
	comment   string
	createdAt time.Time
	updatedAt time.Time
}

func NewExpense(
	userID int64,
	amount Amount,
	category Category,
	spentAt time.Time,
	comment string,
) (*Expense, error) {
	if userID <= 0 {
		return nil, pkgerrs.NewValueInvalidError("user_id")
	}
	if !category.IsValid() {
		return nil, errs.ErrInvalidCategory
	}
	if spentAt.IsZero() {
		return nil, errs.ErrInvalidSpentAt
	}

	now := time.Now().UTC()
	return &Expense{
		id:        uuid.New(),
		userID:    userID,
		amount:    amount,
		category:  category,
		spentAt:   spentAt.UTC(),
		comment:   strings.TrimSpace(comment),
		createdAt: now,
		updatedAt: now,
	}, nil
}

// ================ Read-Only ================

func (e *Expense) ID() uuid.UUID        { return e.id }
func (e *Expense) UserID() int64        { return e.userID }
func (e *Expense) Amount() Amount       { return e.amount }
func (e *Expense) Category() Category   { return e.category }
func (e *Expense) SpentAt() time.Time   { return e.spentAt }
func (e *Expense) Comment() string      { return e.comment }
func (e *Expense) CreatedAt() time.Time { return e.createdAt }
func (e *Expense) UpdatedAt() time.Time { return e.updatedAt }

func (e *Expense) Update(amount Amount, category Category, spentAt time.Time, comment string) error {
	if !category.IsValid() {
		return errs.ErrInvalidCategory
	}
	if spentAt.IsZero() {
		return errs.ErrInvalidSpentAt
	}

	e.amount = amount
	e.category = category
	e.spentAt = spentAt.UTC()
	e.comment = strings.TrimSpace(comment)
	e.updatedAt = time.Now().UTC()
	return nil
}

func ReconstituteExpense(
	id uuid.UUID,
	userID int64,
	amount Amount,
	category Category,
	spentAt time.Time,
	comment string,
	createdAt time.Time,
	updatedAt time.Time,
) *Expense {
	return &Expense{
		id:        id,
		userID:    userID,
		amount:    amount,
		category:  category,
		spentAt:   spentAt,
		comment:   comment,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}
