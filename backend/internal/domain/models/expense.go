package models

import (
	"errors"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	pkgerrs "backend/pkg/errs"
	pkgutils "backend/pkg/utils"

	"github.com/google/uuid"
)

var ErrSpentAtAfterCreatedAt = errors.New("spent_at cannot be after created_at")

type Category string

const (
	CategoryFood          Category = "еда"
	CategoryTransport     Category = "транспорт"
	CategoryHousing       Category = "жильё"
	CategoryEntertainment Category = "развлечения"
	CategoryHealth        Category = "здоровье"
	CategoryShopping      Category = "покупки"
	CategoryOther         Category = "другое"
)

func (c Category) String() string {
	return string(c)
}

func (c Category) IsValid() bool {
	return c == CategoryFood || c == CategoryTransport ||
		c == CategoryHousing || c == CategoryEntertainment ||
		c == CategoryHealth || c == CategoryShopping ||
		c == CategoryOther
}

func NewCategory(val string) (Category, error) {
	c := Category(strings.TrimSpace(val))
	if !c.IsValid() {
		return "", pkgerrs.NewValueInvalidError("category")
	}
	return c, nil
}

const (
	minAmount, maxAmount         = 1.0, 1000000.0
	maxExpenseAge                = time.Hour * 24 * 30
	minCommentLen, maxCommentLen = 2, 256
)

var earliestSpentAt = time.Now().UTC().Add(-maxExpenseAge)

type Expense struct {
	id        uuid.UUID
	userID    int64
	amount    float64
	category  Category
	spentAt   time.Time
	comment   *string
	createdAt time.Time
	updatedAt time.Time
}

func NewExpense(
	userID int64,
	amount float64,
	rawCategory string,
	spentAt time.Time,
	comment *string,
) (*Expense, error) {
	if userID <= 0 {
		return nil, pkgerrs.NewValueInvalidError("user_id")
	}

	roundedAmount := math.Round(amount*100) / 100
	if roundedAmount < minAmount || roundedAmount > maxAmount {
		return nil, pkgerrs.NewValueInvalidError("amount")
	}

	cat := Category(rawCategory)
	if !cat.IsValid() {
		return nil, pkgerrs.NewValueInvalidError("category")
	}

	now := time.Now().UTC()

	if spentAt.IsZero() || spentAt.Before(earliestSpentAt) {
		return nil, pkgerrs.NewValueInvalidError("spent_at")
	}
	if spentAt.After(now) {
		return nil, ErrSpentAtAfterCreatedAt
	}

	var trimmedComm *string
	if comment != nil {
		trimmed := strings.TrimSpace(*comment)
		if trimmed != "" {
			commLen := utf8.RuneCountInString(trimmed)
			if commLen < minCommentLen || commLen > maxCommentLen {
				return nil, pkgerrs.NewValueInvalidError("comment")
			}
			trimmedComm = pkgutils.VPtr(trimmed)
		}
	}

	return &Expense{
		id:        uuid.New(),
		userID:    userID,
		amount:    roundedAmount,
		category:  cat,
		spentAt:   spentAt.UTC(),
		comment:   trimmedComm,
		createdAt: now,
		updatedAt: now,
	}, nil
}

func RestoreExpense(
	id uuid.UUID,
	userID int64,
	amount float64,
	category Category,
	spentAt time.Time,
	comment string,
	createdAt, updatedAt time.Time,
) *Expense {
	var commPtr *string
	trimmed := strings.TrimSpace(comment)
	if trimmed != "" {
		commPtr = pkgutils.VPtr(trimmed)
	}

	return &Expense{
		id:        id,
		userID:    userID,
		amount:    amount,
		category:  category,
		spentAt:   spentAt,
		comment:   commPtr,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func ReconstituteExpense(
	id uuid.UUID,
	userID int64,
	amount float64,
	category Category,
	spentAt time.Time,
	comment string,
	createdAt, updatedAt time.Time,
) *Expense {
	return RestoreExpense(id, userID, amount, category, spentAt, comment, createdAt, updatedAt)
}

func (e *Expense) ID() uuid.UUID        { return e.id }
func (e *Expense) UserID() int64        { return e.userID }
func (e *Expense) Amount() float64      { return e.amount }
func (e *Expense) Category() Category   { return e.category }
func (e *Expense) SpentAt() time.Time   { return e.spentAt }
func (e *Expense) Comment() *string     { return e.comment }
func (e *Expense) CreatedAt() time.Time { return e.createdAt }
func (e *Expense) UpdatedAt() time.Time { return e.updatedAt }

func (e *Expense) Update(
	amount *float64,
	rawCategory *string,
	spentAt *time.Time,
	comment *string,
) error {
	if amount != nil {
		roundedAmount := math.Round(*amount*100) / 100
		if roundedAmount < minAmount || roundedAmount > maxAmount {
			return pkgerrs.NewValueInvalidError("amount")
		}
		e.amount = roundedAmount
	}

	if rawCategory != nil {
		cat := Category(*rawCategory)
		if !cat.IsValid() {
			return pkgerrs.NewValueInvalidError("category")
		}
		e.category = cat
	}

	if spentAt != nil {
		t := *spentAt
		if t.IsZero() || t.Before(earliestSpentAt) {
			return pkgerrs.NewValueInvalidError("spent_at")
		}
		if t.After(e.createdAt) {
			return ErrSpentAtAfterCreatedAt
		}
		e.spentAt = t.UTC()
	}

	if comment != nil {
		trimmed := strings.TrimSpace(*comment)
		if trimmed == "" {
			e.comment = nil
		} else {
			commLen := utf8.RuneCountInString(trimmed)
			if commLen < minCommentLen || commLen > maxCommentLen {
				return pkgerrs.NewValueInvalidError("comment")
			}
			e.comment = pkgutils.VPtr(trimmed)
		}
	}

	e.updatedAt = time.Now().UTC()
	return nil
}
