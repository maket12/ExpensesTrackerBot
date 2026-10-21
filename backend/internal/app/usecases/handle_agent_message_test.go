package usecases

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"backend/internal/adapter/out/memory"
	"backend/internal/domain/models"
	"backend/internal/domain/port"

	"github.com/google/uuid"
)

type mockLLMClient struct {
	responses []port.AgentResponse
	callCount int
}

func (m *mockLLMClient) GenerateResponse(_ context.Context, _ []port.AgentMessage) (port.AgentResponse, error) {
	if m.callCount < len(m.responses) {
		resp := m.responses[m.callCount]
		m.callCount++
		return resp, nil
	}
	return port.AgentResponse{Content: "Готово!"}, nil
}

func TestAgentUseCase_ExecuteTool_UpdateExpense(t *testing.T) {
	ctx := context.Background()
	repo := newMockExpenseRepo()
	expenseUC := NewExpenseUseCase(repo)
	mem := memory.NewStore()
	llm := &mockLLMClient{}
	agent := NewAgentUseCase(expenseUC, llm, mem)

	userID := int64(42)
	e, _ := models.NewExpense(userID, 100, "еда", time.Now().UTC().Add(-time.Hour), "кофе")
	_, _ = repo.Create(ctx, e)

	argsJSON, _ := json.Marshal(map[string]any{
		"id":     e.ID().String(),
		"amount": 200.0,
	})
	res := agent.executeTool(ctx, userID, port.ToolCall{
		ID:        "tc-1",
		Name:      "update_expense",
		Arguments: string(argsJSON),
	})

	if strings.Contains(res, "error") {
		t.Fatalf("unexpected error response: %s", res)
	}

	var updated models.Expense
	expFromRepo, _ := repo.Get(ctx, e.ID())
	if expFromRepo.Amount() != 200.0 {
		t.Errorf("expected amount 200.0, got %f", expFromRepo.Amount())
	}
	if expFromRepo.Category() != "еда" {
		t.Errorf("expected category 'еда', got %s", expFromRepo.Category())
	}
	if expFromRepo.Comment() != "кофе" {
		t.Errorf("expected comment 'кофе', got %s", expFromRepo.Comment())
	}

	argsWrongUser, _ := json.Marshal(map[string]any{
		"id":     e.ID().String(),
		"amount": 500.0,
	})
	resWrongUser := agent.executeTool(ctx, int64(999), port.ToolCall{
		ID:        "tc-2",
		Name:      "update_expense",
		Arguments: string(argsWrongUser),
	})
	if !strings.Contains(resWrongUser, "error") {
		t.Errorf("expected error for wrong user, got: %s", resWrongUser)
	}

	argsInvalidSpentAt, _ := json.Marshal(map[string]any{
		"id":       e.ID().String(),
		"spent_at": "invalid-date",
	})
	resInvalidSpentAt := agent.executeTool(ctx, userID, port.ToolCall{
		ID:        "tc-3",
		Name:      "update_expense",
		Arguments: string(argsInvalidSpentAt),
	})
	if !strings.Contains(resInvalidSpentAt, "error") {
		t.Errorf("expected error for invalid spent_at, got: %s", resInvalidSpentAt)
	}

	resBadJSON := agent.executeTool(ctx, userID, port.ToolCall{
		ID:        "tc-4",
		Name:      "update_expense",
		Arguments: "not-json",
	})
	if !strings.Contains(resBadJSON, "error") {
		t.Errorf("expected error for bad JSON, got: %s", resBadJSON)
	}

	argsNotFound, _ := json.Marshal(map[string]any{
		"id":     uuid.New().String(),
		"amount": 50.0,
	})
	resNotFound := agent.executeTool(ctx, userID, port.ToolCall{
		ID:        "tc-5",
		Name:      "update_expense",
		Arguments: string(argsNotFound),
	})
	if !strings.Contains(resNotFound, "error") {
		t.Errorf("expected error for not found ID, got: %s", resNotFound)
	}
	_ = updated
}
