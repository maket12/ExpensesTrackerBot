package models

import (
	"errors"
	"strings"
	"testing"
	"time"

	pkgerrs "backend/pkg/errs"
	"backend/pkg/utils"

	"github.com/google/uuid"
)

func TestNewExpense_Success(t *testing.T) {
	spentAt := time.Now().UTC().Add(-time.Hour)
	comm := "супермаркет"
	e, err := NewExpense(101, 1500.75, "еда", spentAt, &comm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if e.ID() == uuid.Nil {
		t.Errorf("expected non-nil UUID")
	}
	if e.UserID() != 101 {
		t.Errorf("expected UserID 101, got %d", e.UserID())
	}
	if e.Amount() != 1500.75 {
		t.Errorf("expected Amount 1500.75, got %f", e.Amount())
	}
	if e.Category() != CategoryFood {
		t.Errorf("expected Category food, got %s", e.Category())
	}
	if !e.SpentAt().Equal(spentAt) {
		t.Errorf("expected SpentAt %v, got %v", spentAt, e.SpentAt())
	}
	if e.Comment() == nil || *e.Comment() != comm {
		t.Errorf("expected comment %s, got %v", comm, e.Comment())
	}
	if e.CreatedAt().IsZero() || e.UpdatedAt().IsZero() {
		t.Errorf("expected non-zero CreatedAt and UpdatedAt")
	}
}

func TestNewExpense_NilAndEmptyComment(t *testing.T) {
	spentAt := time.Now().UTC().Add(-10 * time.Minute)

	e1, err := NewExpense(101, 200, "транспорт", spentAt, nil)
	if err != nil {
		t.Fatalf("unexpected error for nil comment: %v", err)
	}
	if e1.Comment() != nil {
		t.Errorf("expected nil comment, got %v", e1.Comment())
	}

	emptyComm := ""
	e2, err := NewExpense(101, 200, "транспорт", spentAt, &emptyComm)
	if err != nil {
		t.Fatalf("unexpected error for empty comment: %v", err)
	}
	if e2.Comment() != nil {
		t.Errorf("expected nil comment for empty string, got %v", e2.Comment())
	}

	spacesComm := "   "
	e3, err := NewExpense(101, 200, "транспорт", spentAt, &spacesComm)
	if err != nil {
		t.Fatalf("unexpected error for whitespace comment: %v", err)
	}
	if e3.Comment() != nil {
		t.Errorf("expected nil comment for whitespace string, got %v", e3.Comment())
	}
}

func TestNewExpense_InvalidUserID(t *testing.T) {
	spentAt := time.Now().UTC().Add(-time.Hour)
	_, err := NewExpense(0, 100, "еда", spentAt, nil)
	if err == nil {
		t.Fatalf("expected error for user_id = 0")
	}
	var valErr *pkgerrs.ValueInvalidError
	if !errors.As(err, &valErr) || valErr.ParamName != "user_id" {
		t.Errorf("expected ValueInvalidError for user_id, got %v", err)
	}
}

func TestNewExpense_InvalidAmount(t *testing.T) {
	spentAt := time.Now().UTC().Add(-time.Hour)

	_, err := NewExpense(101, 0, "еда", spentAt, nil)
	if err == nil {
		t.Fatalf("expected error for amount = 0")
	}

	_, err = NewExpense(101, -50, "еда", spentAt, nil)
	if err == nil {
		t.Fatalf("expected error for negative amount")
	}

	_, err = NewExpense(101, 0.5, "еда", spentAt, nil)
	if err == nil {
		t.Fatalf("expected error for amount < minAmount (1)")
	}

	_, err = NewExpense(101, 1000001, "еда", spentAt, nil)
	if err == nil {
		t.Fatalf("expected error for amount > maxAmount (1000000)")
	}
}

func TestNewExpense_InvalidCategory(t *testing.T) {
	spentAt := time.Now().UTC().Add(-time.Hour)
	_, err := NewExpense(101, 100, "неизвестная_категория", spentAt, nil)
	if err == nil {
		t.Fatalf("expected error for invalid category")
	}
	var valErr *pkgerrs.ValueInvalidError
	if !errors.As(err, &valErr) || valErr.ParamName != "category" {
		t.Errorf("expected ValueInvalidError for category, got %v", err)
	}
}

func TestNewExpense_SpentAtValidation(t *testing.T) {
	_, err := NewExpense(101, 100, "еда", time.Time{}, nil)
	if err == nil {
		t.Fatalf("expected error for zero spentAt")
	}

	futureSpentAt := time.Now().UTC().Add(2 * time.Hour)
	_, err = NewExpense(101, 100, "еда", futureSpentAt, nil)
	if err == nil {
		t.Fatalf("expected error for future spentAt (spentAt > createdAt)")
	}

	tooOldSpentAt := earliestSpentAt.Add(-2 * time.Hour)
	_, err = NewExpense(101, 100, "еда", tooOldSpentAt, nil)
	if err == nil {
		t.Fatalf("expected error for spentAt earlier than earliestSpentAt")
	}

	validPastSpentAt := earliestSpentAt.Add(2 * time.Hour)
	e, err := NewExpense(101, 100, "еда", validPastSpentAt, nil)
	if err != nil {
		t.Fatalf("unexpected error for valid spentAt: %v", err)
	}
	if !e.SpentAt().Equal(validPastSpentAt) {
		t.Errorf("expected spentAt %v, got %v", validPastSpentAt, e.SpentAt())
	}
}

func TestNewExpense_CommentLengthValidation(t *testing.T) {
	spentAt := time.Now().UTC().Add(-time.Hour)

	shortComm := "a"
	_, err := NewExpense(101, 100, "еда", spentAt, &shortComm)
	if err == nil {
		t.Fatalf("expected error for comment shorter than 2 chars")
	}

	longComm := strings.Repeat("x", 257)
	_, err = NewExpense(101, 100, "еда", spentAt, &longComm)
	if err == nil {
		t.Fatalf("expected error for comment longer than 256 chars")
	}

	validComm := "ок"
	e, err := NewExpense(101, 100, "еда", spentAt, &validComm)
	if err != nil {
		t.Fatalf("unexpected error for 2-char comment: %v", err)
	}
	if *e.Comment() != "ок" {
		t.Errorf("expected comment 'ок', got %s", *e.Comment())
	}
}

func TestExpense_Update(t *testing.T) {
	spentAt := time.Now().UTC().Add(-time.Hour)
	e, err := NewExpense(101, 100, "еда", spentAt, utils.VPtr("обед"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	newSpentAt := time.Now().UTC().Add(-30 * time.Minute)
	err = e.Update(250.50, "транспорт", newSpentAt, utils.VPtr("метро"))
	if err != nil {
		t.Fatalf("unexpected error on valid update: %v", err)
	}

	if e.Amount() != 250.50 {
		t.Errorf("expected Amount 250.50, got %f", e.Amount())
	}
	if e.Category() != CategoryTransport {
		t.Errorf("expected Category transport, got %s", e.Category())
	}
	if e.Comment() == nil || *e.Comment() != "метро" {
		t.Errorf("expected comment 'метро', got %v", e.Comment())
	}

	err = e.Update(0, "транспорт", newSpentAt, nil)
	if err == nil {
		t.Errorf("expected error updating with invalid amount")
	}

	err = e.Update(100, "неверная", newSpentAt, nil)
	if err == nil {
		t.Errorf("expected error updating with invalid category")
	}

	err = e.Update(100, "еда", time.Now().UTC().Add(time.Hour), nil)
	if err == nil {
		t.Errorf("expected error updating with future spentAt")
	}
}

func TestRestoreExpense(t *testing.T) {
	id := uuid.New()
	spentAt := time.Now().UTC().Add(-2 * time.Hour)
	createdAt := time.Now().UTC().Add(-time.Hour)
	updatedAt := time.Now().UTC()

	e := RestoreExpense(id, 202, 350.25, CategoryHealth, spentAt, "аптека", createdAt, updatedAt)
	if e.ID() != id {
		t.Errorf("expected ID %v, got %v", id, e.ID())
	}
	if e.UserID() != 202 {
		t.Errorf("expected UserID 202, got %d", e.UserID())
	}
	if e.Amount() != 350.25 {
		t.Errorf("expected Amount 350.25, got %f", e.Amount())
	}
	if e.Category() != CategoryHealth {
		t.Errorf("expected Category health, got %s", e.Category())
	}
	if e.Comment() == nil || *e.Comment() != "аптека" {
		t.Errorf("expected comment 'аптека', got %v", e.Comment())
	}

	eEmpty := RestoreExpense(id, 202, 350.25, CategoryHealth, spentAt, "", createdAt, updatedAt)
	if eEmpty.Comment() != nil {
		t.Errorf("expected nil comment when empty string restored, got %v", eEmpty.Comment())
	}
}

func TestCategory_Validation(t *testing.T) {
	c, err := NewCategory("еда")
	if err != nil || c != CategoryFood {
		t.Errorf("expected CategoryFood, got %v, err: %v", c, err)
	}

	_, err = NewCategory("неизвестная")
	if err == nil {
		t.Errorf("expected error for invalid category string")
	}
}
