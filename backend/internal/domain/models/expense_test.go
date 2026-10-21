package models

import (
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	pkgerrs "backend/pkg/errs"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
)

func randomValidCategory() string {
	categories := []string{"еда", "транспорт", "жильё", "развлечения", "здоровье", "покупки", "другое"}
	return gofakeit.RandomString(categories)
}

func TestNewExpense_WithGoFakeIt(t *testing.T) {
	userID := int64(gofakeit.Number(1, 100000))
	amount := gofakeit.Float64Range(1.0, 10000.0)
	cat := randomValidCategory()
	spentAt := time.Now().UTC().Add(-time.Duration(gofakeit.Number(1, 100)) * time.Minute)
	comm := gofakeit.Sentence(3)

	e, err := NewExpense(userID, amount, cat, spentAt, comm)
	if err != nil {
		t.Fatalf("unexpected error creating expense with fake data: %v", err)
	}

	if e.ID() == uuid.Nil {
		t.Errorf("expected generated UUID")
	}
	if e.UserID() != userID {
		t.Errorf("expected userID %d, got %d", userID, e.UserID())
	}
	if e.Amount() <= 0 {
		t.Errorf("expected positive amount, got %f", e.Amount())
	}
	if string(e.Category()) != cat {
		t.Errorf("expected category %s, got %s", cat, e.Category())
	}
	if e.Comment() != comm {
		t.Errorf("expected comment %s, got %v", comm, e.Comment())
	}
}

func TestNewExpense_OptionalComment(t *testing.T) {
	userID := int64(gofakeit.Number(1, 100000))
	amount := gofakeit.Float64Range(1.0, 10000.0)
	cat := randomValidCategory()
	spentAt := time.Now().UTC().Add(-time.Hour)

	e, err := NewExpense(userID, amount, cat, spentAt, "")
	if err != nil {
		t.Fatalf("unexpected error creating expense without comment: %v", err)
	}
	if e.Comment() != "" {
		t.Errorf("expected empty comment, got %q", e.Comment())
	}
}

func TestNewExpense_SpentAtAfterCreatedAtError(t *testing.T) {
	userID := int64(gofakeit.Number(1, 1000))
	amount := gofakeit.Float64Range(10.0, 500.0)
	cat := randomValidCategory()
	futureSpentAt := time.Now().UTC().Add(time.Hour)

	_, err := NewExpense(userID, amount, cat, futureSpentAt, "")
	if err == nil {
		t.Fatalf("expected error when spentAt > createdAt")
	}
	if !errors.Is(err, ErrSpentAtAfterCreatedAt) {
		t.Errorf("expected ErrSpentAtAfterCreatedAt, got %v", err)
	}
}

func TestNewExpense_SpentAtTooOld(t *testing.T) {
	userID := int64(gofakeit.Number(1, 1000))
	amount := gofakeit.Float64Range(10.0, 500.0)
	cat := randomValidCategory()
	tooOldSpentAt := earliestSpentAt.Add(-time.Hour)

	_, err := NewExpense(userID, amount, cat, tooOldSpentAt, "")
	if err == nil {
		t.Fatalf("expected error for spentAt before earliestSpentAt")
	}
	var valErr *pkgerrs.ValueInvalidError
	if !errors.As(err, &valErr) || valErr.ParamName != "spent_at" {
		t.Errorf("expected ValueInvalidError for spent_at, got %v", err)
	}
}

func TestNewExpense_ValidationErrors(t *testing.T) {
	validSpentAt := time.Now().UTC().Add(-time.Minute)

	_, err := NewExpense(0, 100, "еда", validSpentAt, "")
	if err == nil {
		t.Errorf("expected error for user_id <= 0")
	}

	_, err = NewExpense(10, 0.5, "еда", validSpentAt, "")
	if err == nil {
		t.Errorf("expected error for amount < minAmount (1)")
	}

	_, err = NewExpense(10, 1000001, "еда", validSpentAt, "")
	if err == nil {
		t.Errorf("expected error for amount > maxAmount (1000000)")
	}

	_, err = NewExpense(10, 100, "несуществующая", validSpentAt, "")
	if err == nil {
		t.Errorf("expected error for invalid category")
	}

	shortComm := "x"
	_, err = NewExpense(10, 100, "еда", validSpentAt, shortComm)
	if err == nil {
		t.Errorf("expected error for comment len < 2")
	}

	longComm := strings.Repeat("x", 257)
	_, err = NewExpense(10, 100, "еда", validSpentAt, longComm)
	if err == nil {
		t.Errorf("expected error for comment len > 256")
	}

	whitespaceComm := "   "
	_, err = NewExpense(10, 100, "еда", validSpentAt, whitespaceComm)
	if err == nil {
		t.Errorf("expected error for whitespace comment")
	}
}

func TestExpense_Update_PatchSemantics(t *testing.T) {
	userID := int64(gofakeit.Number(1, 1000))
	initialAmount := 100.0
	initialSpentAt := time.Now().UTC().Add(-2 * time.Hour)
	initialComment := gofakeit.Sentence(2)

	e, err := NewExpense(userID, initialAmount, "еда", initialSpentAt, initialComment)
	if err != nil {
		t.Fatalf("unexpected error creating expense: %v", err)
	}

	err = e.Update(nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error updating zero fields: %v", err)
	}
	if e.Amount() != initialAmount {
		t.Errorf("expected amount %f to remain unchanged, got %f", initialAmount, e.Amount())
	}
	if e.Category() != CategoryFood {
		t.Errorf("expected category food to remain unchanged")
	}
	if e.Comment() != initialComment {
		t.Errorf("expected comment %s to remain unchanged", initialComment)
	}

	newAmount := math.Round(gofakeit.Float64Range(200.0, 300.0)*100) / 100
	err = e.Update(&newAmount, nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error updating amount only: %v", err)
	}
	if e.Amount() != newAmount {
		t.Errorf("expected amount updated to %f, got %f", newAmount, e.Amount())
	}

	newCat := "транспорт"
	err = e.Update(nil, &newCat, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error updating category only: %v", err)
	}
	if e.Category() != CategoryTransport {
		t.Errorf("expected category updated to %s, got %s", newCat, e.Category())
	}

	newSpentAt := time.Now().UTC().Add(-time.Hour)
	err = e.Update(nil, nil, &newSpentAt, nil)
	if err != nil {
		t.Fatalf("unexpected error updating spentAt only: %v", err)
	}
	if !e.SpentAt().Equal(newSpentAt) {
		t.Errorf("expected spentAt updated to %v, got %v", newSpentAt, e.SpentAt())
	}

	newComment := gofakeit.Sentence(2)
	err = e.Update(nil, nil, nil, &newComment)
	if err != nil {
		t.Fatalf("unexpected error updating comment only: %v", err)
	}
	if e.Comment() != newComment {
		t.Errorf("expected comment updated to %s, got %v", newComment, e.Comment())
	}

	clearComment := ""
	err = e.Update(nil, nil, nil, &clearComment)
	if err != nil {
		t.Fatalf("unexpected error clearing comment: %v", err)
	}
	if e.Comment() != "" {
		t.Errorf("expected comment to be cleared to empty string, got %v", e.Comment())
	}
}

func TestExpense_Update_ValidationFailures(t *testing.T) {
	e, err := NewExpense(10, 100, "еда", time.Now().UTC().Add(-time.Hour), "обед")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	invalidAmount := 0.0
	err = e.Update(&invalidAmount, nil, nil, nil)
	if err == nil {
		t.Errorf("expected error for amount = 0")
	}

	invalidCat := "неизвестная"
	err = e.Update(nil, &invalidCat, nil, nil)
	if err == nil {
		t.Errorf("expected error for invalid category")
	}

	futureSpentAt := time.Now().UTC().Add(time.Hour)
	err = e.Update(nil, nil, &futureSpentAt, nil)
	if err == nil {
		t.Errorf("expected error for spentAt > createdAt")
	}
	if !errors.Is(err, ErrSpentAtAfterCreatedAt) {
		t.Errorf("expected ErrSpentAtAfterCreatedAt, got %v", err)
	}

	oldSpentAt := earliestSpentAt.Add(-2 * time.Hour)
	err = e.Update(nil, nil, &oldSpentAt, nil)
	if err == nil {
		t.Errorf("expected error for spentAt before earliestSpentAt")
	}

	shortComm := "a"
	err = e.Update(nil, nil, nil, &shortComm)
	if err == nil {
		t.Errorf("expected error for comment < 2 runes")
	}

	longComm := strings.Repeat("x", 257)
	err = e.Update(nil, nil, nil, &longComm)
	if err == nil {
		t.Errorf("expected error for comment > 256 runes")
	}

	whitespaceComm := "   "
	err = e.Update(nil, nil, nil, &whitespaceComm)
	if err == nil {
		t.Errorf("expected error for whitespace comment")
	}
}

func TestRestoreExpense_WithGoFakeIt(t *testing.T) {
	id := uuid.New()
	userID := int64(gofakeit.Number(1, 1000))
	amount := gofakeit.Float64Range(10.0, 500.0)
	cat := CategoryFood
	spentAt := time.Now().UTC().Add(-2 * time.Hour)
	createdAt := time.Now().UTC().Add(-time.Hour)
	updatedAt := time.Now().UTC()
	comm := gofakeit.Word() + " " + gofakeit.Word()

	e := RestoreExpense(id, userID, amount, cat, spentAt, comm, createdAt, updatedAt)
	if e.ID() != id || e.UserID() != userID || e.Amount() != amount || e.Category() != cat {
		t.Errorf("restored fields do not match")
	}
	if e.Comment() != comm {
		t.Errorf("expected comment %s, got %v", comm, e.Comment())
	}
}
