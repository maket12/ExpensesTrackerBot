package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	"backend/internal/app/dto"
	apperrs "backend/internal/app/errs"
	"backend/internal/domain/models"
	"backend/internal/domain/port"
	pkgerrs "backend/pkg/errs"

	"github.com/google/uuid"
)

type mockExpenseRepo struct {
	expenses map[uuid.UUID]*models.Expense
}

func newMockExpenseRepo() *mockExpenseRepo {
	return &mockExpenseRepo{
		expenses: make(map[uuid.UUID]*models.Expense),
	}
}

func (m *mockExpenseRepo) Create(_ context.Context, expense *models.Expense) (*models.Expense, error) {
	m.expenses[expense.ID()] = expense
	return expense, nil
}

func (m *mockExpenseRepo) Get(_ context.Context, id uuid.UUID) (*models.Expense, error) {
	expense, ok := m.expenses[id]
	if !ok {
		return nil, pkgerrs.NewObjectNotFoundError("expense", id)
	}
	return expense, nil
}

func (m *mockExpenseRepo) Update(_ context.Context, expense *models.Expense) (*models.Expense, error) {
	if _, ok := m.expenses[expense.ID()]; !ok {
		return nil, pkgerrs.NewObjectNotFoundError("expense", expense.ID())
	}
	m.expenses[expense.ID()] = expense
	return expense, nil
}

func (m *mockExpenseRepo) Delete(_ context.Context, id uuid.UUID) error {
	if _, ok := m.expenses[id]; !ok {
		return pkgerrs.NewObjectNotFoundError("expense", id)
	}
	delete(m.expenses, id)
	return nil
}

func (m *mockExpenseRepo) List(_ context.Context, filter port.ExpenseFilter) ([]*models.Expense, error) {
	var res []*models.Expense
	for _, e := range m.expenses {
		if filter.UserID != nil && e.UserID() != *filter.UserID {
			continue
		}
		if filter.Category != nil && e.Category() != *filter.Category {
			continue
		}
		if filter.From != nil && e.SpentAt().Before(*filter.From) {
			continue
		}
		if filter.To != nil && e.SpentAt().After(*filter.To) {
			continue
		}
		res = append(res, e)
	}
	return res, nil
}

func (m *mockExpenseRepo) ListLatest(_ context.Context, userID int64, limit int) ([]*models.Expense, error) {
	var res []*models.Expense
	for _, e := range m.expenses {
		if e.UserID() == userID {
			res = append(res, e)
			if limit > 0 && len(res) >= limit {
				break
			}
		}
	}
	return res, nil
}

func (m *mockExpenseRepo) ListUserIDS(_ context.Context) ([]int64, error) {
	set := make(map[int64]struct{})
	for _, e := range m.expenses {
		set[e.UserID()] = struct{}{}
	}
	var res []int64
	for id := range set {
		res = append(res, id)
	}
	return res, nil
}

func TestCreateExpenseUC(t *testing.T) {
	ctx := context.Background()
	repo := newMockExpenseRepo()
	uc := NewCreateExpenseUC(repo)

	spentAt := time.Now().UTC().Add(-time.Hour)
	out, err := uc.Execute(ctx, dto.CreateExpenseInput{
		UserID:   100,
		Amount:   250.50,
		Category: "еда",
		SpentAt:  spentAt,
		Comment:  "обед в кафе",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.Expense.ID == "" {
		t.Errorf("expected expense ID to be set")
	}
	if out.Expense.Amount != 250.50 {
		t.Errorf("expected amount 250.50, got %f", out.Expense.Amount)
	}
	if out.Expense.Category != "еда" {
		t.Errorf("expected category 'еда', got %s", out.Expense.Category)
	}

	_, err = uc.Execute(ctx, dto.CreateExpenseInput{
		UserID:   100,
		Amount:   0.5,
		Category: "еда",
		SpentAt:  spentAt,
	})
	if err == nil {
		t.Errorf("expected validation error for amount < 1.0")
	}

	_, err = uc.Execute(ctx, dto.CreateExpenseInput{
		UserID:   100,
		Amount:   100,
		Category: "неизвестно",
		SpentAt:  spentAt,
	})
	if err == nil {
		t.Errorf("expected validation error for invalid category")
	}
}

func TestGetExpenseUC(t *testing.T) {
	ctx := context.Background()
	repo := newMockExpenseRepo()
	uc := NewGetExpenseUC(repo)

	e, _ := models.NewExpense(100, 150, "еда", time.Now().UTC().Add(-time.Minute), "кофе")
	_, _ = repo.Create(ctx, e)

	out, err := uc.Execute(ctx, dto.GetExpenseInput{ID: e.ID().String()})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Expense.ID != e.ID().String() {
		t.Errorf("expected ID %s, got %s", e.ID().String(), out.Expense.ID)
	}

	_, err = uc.Execute(ctx, dto.GetExpenseInput{ID: "invalid-uuid"})
	if err == nil || !errors.Is(err, apperrs.ErrInvalidInput) {
		t.Errorf("expected invalid input error, got %v", err)
	}

	_, err = uc.Execute(ctx, dto.GetExpenseInput{ID: uuid.New().String()})
	if err == nil || !errors.Is(err, apperrs.ErrExpenseNotFound) {
		t.Errorf("expected not found error, got %v", err)
	}
}

func TestUpdateExpenseUC_PatchSemantics(t *testing.T) {
	ctx := context.Background()
	repo := newMockExpenseRepo()
	uc := NewUpdateExpenseUC(repo)

	initialSpentAt := time.Now().UTC().Add(-2 * time.Hour)
	e, _ := models.NewExpense(100, 300, "еда", initialSpentAt, "пицца")
	_, _ = repo.Create(ctx, e)

	newAmount := 450.0
	out, err := uc.Execute(ctx, dto.UpdateExpenseInput{
		ID:     e.ID().String(),
		Amount: &newAmount,
	})
	if err != nil {
		t.Fatalf("unexpected error updating amount: %v", err)
	}
	if out.Expense.Amount != 450.0 {
		t.Errorf("expected amount 450.0, got %f", out.Expense.Amount)
	}
	if out.Expense.Category != "еда" {
		t.Errorf("expected category to remain 'еда', got %s", out.Expense.Category)
	}
	if out.Expense.Comment != "пицца" {
		t.Errorf("expected comment to remain 'пицца', got %s", out.Expense.Comment)
	}

	newCat := "развлечения"
	out, err = uc.Execute(ctx, dto.UpdateExpenseInput{
		ID:       e.ID().String(),
		Category: &newCat,
	})
	if err != nil {
		t.Fatalf("unexpected error updating category: %v", err)
	}
	if out.Expense.Category != "развлечения" {
		t.Errorf("expected category 'развлечения', got %s", out.Expense.Category)
	}

	newSpentAt := time.Now().UTC().Add(-time.Hour)
	out, err = uc.Execute(ctx, dto.UpdateExpenseInput{
		ID:      e.ID().String(),
		SpentAt: &newSpentAt,
	})
	if err != nil {
		t.Fatalf("unexpected error updating spentAt: %v", err)
	}
	if !out.Expense.SpentAt.Equal(newSpentAt.UTC()) {
		t.Errorf("expected spentAt %v, got %v", newSpentAt.UTC(), out.Expense.SpentAt)
	}

	clearComment := ""
	out, err = uc.Execute(ctx, dto.UpdateExpenseInput{
		ID:      e.ID().String(),
		Comment: &clearComment,
	})
	if err != nil {
		t.Fatalf("unexpected error clearing comment: %v", err)
	}
	if out.Expense.Comment != "" {
		t.Errorf("expected comment to be cleared, got %s", out.Expense.Comment)
	}
}

func TestUpdateExpenseUC_ValidationErrors(t *testing.T) {
	ctx := context.Background()
	repo := newMockExpenseRepo()
	uc := NewUpdateExpenseUC(repo)

	e, _ := models.NewExpense(100, 300, "еда", time.Now().UTC().Add(-time.Hour), "пицца")
	_, _ = repo.Create(ctx, e)

	invalidAmount := 0.0
	_, err := uc.Execute(ctx, dto.UpdateExpenseInput{
		ID:     e.ID().String(),
		Amount: &invalidAmount,
	})
	if err == nil {
		t.Errorf("expected domain error for amount = 0")
	}

	invalidCat := "абракадабра"
	_, err = uc.Execute(ctx, dto.UpdateExpenseInput{
		ID:       e.ID().String(),
		Category: &invalidCat,
	})
	if err == nil {
		t.Errorf("expected domain error for invalid category")
	}

	futureSpentAt := time.Now().UTC().Add(time.Hour)
	_, err = uc.Execute(ctx, dto.UpdateExpenseInput{
		ID:      e.ID().String(),
		SpentAt: &futureSpentAt,
	})
	if err == nil || !errors.Is(err, models.ErrSpentAtAfterCreatedAt) {
		t.Errorf("expected ErrSpentAtAfterCreatedAt, got %v", err)
	}
}

func TestUpdateExpenseUC_EnforceUserID(t *testing.T) {
	ctx := context.Background()
	repo := newMockExpenseRepo()
	uc := NewUpdateExpenseUC(repo)

	e, _ := models.NewExpense(100, 300, "еда", time.Now().UTC().Add(-time.Hour), "пицца")
	_, _ = repo.Create(ctx, e)

	newAmount := 500.0
	out, err := uc.Execute(ctx, dto.UpdateExpenseInput{
		ID:            e.ID().String(),
		UserID:        100,
		Amount:        &newAmount,
		EnforceUserID: true,
	})
	if err != nil {
		t.Fatalf("unexpected error with correct user: %v", err)
	}
	if out.Expense.Amount != 500.0 {
		t.Errorf("expected amount 500.0, got %f", out.Expense.Amount)
	}

	_, err = uc.Execute(ctx, dto.UpdateExpenseInput{
		ID:            e.ID().String(),
		UserID:        200,
		Amount:        &newAmount,
		EnforceUserID: true,
	})
	if err == nil || !errors.Is(err, apperrs.ErrForbiddenAccess) {
		t.Errorf("expected ErrForbiddenAccess, got %v", err)
	}
}

func TestDeleteExpenseUC(t *testing.T) {
	ctx := context.Background()
	repo := newMockExpenseRepo()
	uc := NewDeleteExpenseUC(repo)

	e, _ := models.NewExpense(100, 300, "еда", time.Now().UTC().Add(-time.Hour), "пицца")
	_, _ = repo.Create(ctx, e)

	_, err := uc.Execute(ctx, dto.DeleteExpenseInput{
		ID:            e.ID().String(),
		UserID:        200,
		EnforceUserID: true,
	})
	if err == nil || !errors.Is(err, apperrs.ErrForbiddenAccess) {
		t.Errorf("expected ErrForbiddenAccess, got %v", err)
	}

	out, err := uc.Execute(ctx, dto.DeleteExpenseInput{
		ID:            e.ID().String(),
		UserID:        100,
		EnforceUserID: true,
	})
	if err != nil || !out.Success {
		t.Fatalf("unexpected error deleting: %v", err)
	}
}

func TestListExpensesUC(t *testing.T) {
	ctx := context.Background()
	repo := newMockExpenseRepo()
	uc := NewListExpensesUC(repo)

	e1, _ := models.NewExpense(100, 100, "еда", time.Now().UTC().Add(-time.Hour), "кофе")
	e2, _ := models.NewExpense(200, 200, "транспорт", time.Now().UTC().Add(-30*time.Minute), "метро")
	_, _ = repo.Create(ctx, e1)
	_, _ = repo.Create(ctx, e2)

	userID := int64(100)
	out, err := uc.Execute(ctx, dto.ListExpensesInput{
		UserID: &userID,
	})
	if err != nil {
		t.Fatalf("unexpected error listing: %v", err)
	}
	if len(out.Expenses) != 1 {
		t.Errorf("expected 1 expense, got %d", len(out.Expenses))
	}

	invalidCat := "невалидно"
	_, err = uc.Execute(ctx, dto.ListExpensesInput{
		Category: &invalidCat,
	})
	if err == nil {
		t.Errorf("expected error for invalid category in filter")
	}
}

func TestGetDistinctUsersUC(t *testing.T) {
	ctx := context.Background()
	repo := newMockExpenseRepo()
	uc := NewGetDistinctUsersUC(repo)

	e1, _ := models.NewExpense(100, 100, "еда", time.Now().UTC().Add(-time.Hour), "кофе")
	e2, _ := models.NewExpense(200, 200, "транспорт", time.Now().UTC().Add(-30*time.Minute), "метро")
	_, _ = repo.Create(ctx, e1)
	_, _ = repo.Create(ctx, e2)

	out, err := uc.Execute(ctx, dto.GetDistinctUsersInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.UserIDs) != 2 {
		t.Errorf("expected 2 user IDs, got %d", len(out.UserIDs))
	}
}
