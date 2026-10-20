package mapper

import (
	"testing"
	"time"

	"backend/internal/adapter/out/postgres/sqlc"
	"backend/internal/domain/models"
	"backend/pkg/utils"

	"github.com/google/uuid"
)

func TestMapperUUID(t *testing.T) {
	id := uuid.New()
	pgUUID := ToPgUUID(id)
	if !pgUUID.Valid {
		t.Fatalf("expected pgUUID to be valid")
	}

	converted := FromPgUUID(pgUUID)
	if converted != id {
		t.Fatalf("expected %v, got %v", id, converted)
	}
}

func TestMapperNumeric(t *testing.T) {
	val := 250.75
	pgNum := ToPgNumeric(val)
	if !pgNum.Valid {
		t.Fatalf("expected pgNum to be valid")
	}

	got, err := FromPgNumeric(pgNum)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != val {
		t.Fatalf("expected %v, got %v", val, got)
	}
}

func TestMapperTimestamptz(t *testing.T) {
	now := time.Now().Truncate(time.Microsecond)
	pgTime := ToPgTimestamptz(now)
	if !pgTime.Valid {
		t.Fatalf("expected pgTime to be valid")
	}

	got := FromPgTimestamptz(pgTime)
	if !got.Equal(now) {
		t.Fatalf("expected %v, got %v", now, got)
	}

	zeroTime := ToPgTimestamptz(time.Time{})
	if zeroTime.Valid {
		t.Fatalf("expected zero time to be invalid")
	}
}

func TestMapperParamsAndDomain(t *testing.T) {
	amount, err := models.NewAmount(120.50)
	if err != nil {
		t.Fatalf("unexpected error creating amount: %v", err)
	}

	cat, err := models.NewCategory("еда")
	if err != nil {
		t.Fatalf("unexpected error creating category: %v", err)
	}

	spentAt := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	expense, err := models.NewExpense(12345, amount, cat, spentAt, "обед")
	if err != nil {
		t.Fatalf("unexpected error creating expense: %v", err)
	}

	createParams := ToCreateExpenseParams(expense)
	if createParams.UserID != 12345 {
		t.Errorf("expected UserID 12345, got %d", createParams.UserID)
	}
	if createParams.Category != "еда" {
		t.Errorf("expected category 'еда', got %s", createParams.Category)
	}

	updateParams := ToUpdateExpenseParams(expense)
	if updateParams.Category != "еда" {
		t.Errorf("expected category 'еда', got %s", updateParams.Category)
	}

	filter := models.ExpenseFilter{
		UserID:   utils.VPtr(int64(12345)),
		Category: utils.VPtr(models.CategoryFood),
		From:     utils.VPtr(spentAt.Add(-time.Hour)),
		To:       utils.VPtr(spentAt.Add(time.Hour)),
		Limit:    10,
		Offset:   0,
	}

	listParams := ToListExpensesFilteredParams(filter)
	if *listParams.UserID != 12345 {
		t.Errorf("expected UserID 12345, got %d", *listParams.UserID)
	}
	if *listParams.Category != "еда" {
		t.Errorf("expected category 'еда', got %s", *listParams.Category)
	}
	if !listParams.From.Valid || !listParams.To.Valid {
		t.Errorf("expected From and To to be valid")
	}

	row := sqlc.Expense{
		ID:        createParams.ID,
		UserID:    createParams.UserID,
		Amount:    createParams.Amount,
		Category:  createParams.Category,
		SpentAt:   createParams.SpentAt,
		Comment:   createParams.Comment,
		CreatedAt: createParams.CreatedAt,
		UpdatedAt: createParams.UpdatedAt,
	}

	domainExpense, err := ToDomainExpense(row)
	if err != nil {
		t.Fatalf("unexpected error converting to domain: %v", err)
	}

	if domainExpense.ID() != expense.ID() {
		t.Errorf("expected ID %v, got %v", expense.ID(), domainExpense.ID())
	}
	if domainExpense.Amount().Value() != 120.50 {
		t.Errorf("expected amount 120.50, got %f", domainExpense.Amount().Value())
	}
	if domainExpense.Category() != models.CategoryFood {
		t.Errorf("expected category %v, got %v", models.CategoryFood, domainExpense.Category())
	}
}
