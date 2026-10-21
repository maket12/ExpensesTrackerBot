package usecases

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"backend/internal/adapter/out/memory"
	"backend/internal/app/dto"
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

func TestBuildAgentSystemPrompt(t *testing.T) {
	now := time.Date(2026, 10, 21, 14, 0, 0, 0, time.UTC)
	prompt := BuildAgentSystemPrompt(now)
	if !strings.Contains(prompt, "Среда") {
		t.Errorf("expected prompt to contain weekday 'Среда'")
	}
	if !strings.Contains(prompt, "21.10.2026") {
		t.Errorf("expected prompt to contain date '21.10.2026'")
	}
}

func TestAgentToolExecutor_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := newMockExpenseRepo()
	createUC := NewCreateExpenseUC(repo)
	listUC := NewListExpensesUC(repo)
	updateUC := NewUpdateExpenseUC(repo)
	deleteUC := NewDeleteExpenseUC(repo)
	executor := NewAgentToolExecutor(createUC, listUC, updateUC, deleteUC)

	userID := int64(42)

	addJSON, _ := json.Marshal(map[string]any{
		"amount":   350.0,
		"category": "еда",
		"spent_at": time.Now().UTC().Add(-time.Hour).Format(time.RFC3339),
		"comment":  "обед",
	})
	resAdd := executor.Execute(ctx, userID, port.ToolCall{
		ID:        "tc-1",
		Name:      "add_expense",
		Arguments: string(addJSON),
	})
	if strings.Contains(resAdd, "error") {
		t.Fatalf("unexpected add_expense error: %s", resAdd)
	}

	var added dto.ExpenseDTO
	_ = json.Unmarshal([]byte(resAdd), &added)
	if added.Amount != 350.0 {
		t.Errorf("expected amount 350.0, got %f", added.Amount)
	}

	listJSON, _ := json.Marshal(map[string]any{
		"limit": 10,
	})
	resList := executor.Execute(ctx, userID, port.ToolCall{
		ID:        "tc-2",
		Name:      "list_expenses",
		Arguments: string(listJSON),
	})
	if strings.Contains(resList, "error") {
		t.Fatalf("unexpected list_expenses error: %s", resList)
	}
	if !strings.Contains(resList, "350") {
		t.Errorf("expected list to contain 350: %s", resList)
	}

	updateJSON, _ := json.Marshal(map[string]any{
		"id":     added.ID,
		"amount": 400.0,
	})
	resUpdate := executor.Execute(ctx, userID, port.ToolCall{
		ID:        "tc-3",
		Name:      "update_expense",
		Arguments: string(updateJSON),
	})
	if strings.Contains(resUpdate, "error") {
		t.Fatalf("unexpected update_expense error: %s", resUpdate)
	}
	expFromRepo, _ := repo.Get(ctx, uuid.MustParse(added.ID))
	if expFromRepo.Amount() != 400.0 {
		t.Errorf("expected updated amount 400.0, got %f", expFromRepo.Amount())
	}

	deleteJSON, _ := json.Marshal(map[string]any{
		"id": added.ID,
	})
	resDelete := executor.Execute(ctx, userID, port.ToolCall{
		ID:        "tc-4",
		Name:      "delete_expense",
		Arguments: string(deleteJSON),
	})
	if strings.Contains(resDelete, "error") {
		t.Fatalf("unexpected delete_expense error: %s", resDelete)
	}

	unknownRes := executor.Execute(ctx, userID, port.ToolCall{
		ID:   "tc-5",
		Name: "unknown_tool",
	})
	if !strings.Contains(unknownRes, "error") {
		t.Errorf("expected error for unknown tool")
	}
}

func TestHandleAgentMessageUC_Execute(t *testing.T) {
	ctx := context.Background()
	repo := newMockExpenseRepo()
	createUC := NewCreateExpenseUC(repo)
	listUC := NewListExpensesUC(repo)
	updateUC := NewUpdateExpenseUC(repo)
	deleteUC := NewDeleteExpenseUC(repo)
	executor := NewAgentToolExecutor(createUC, listUC, updateUC, deleteUC)

	mem := memory.NewStore()

	addJSON, _ := json.Marshal(map[string]any{
		"amount":   150.0,
		"category": "еда",
		"spent_at": time.Now().UTC().Add(-time.Hour).Format(time.RFC3339),
		"comment":  "кофе",
	})
	llm := &mockLLMClient{
		responses: []port.AgentResponse{
			{
				ToolCalls: []port.ToolCall{
					{
						ID:        "call-1",
						Name:      "add_expense",
						Arguments: string(addJSON),
					},
				},
			},
			{
				Content: "Записал: кофе 150 ₽ (еда)",
			},
		},
	}

	uc := NewHandleAgentMessageUC(llm, mem, executor)

	out, err := uc.Execute(ctx, dto.HandleAgentMessageInput{
		UserID: 42,
		Text:   "кофе 150",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.ResponseText != "Записал: кофе 150 ₽ (еда)" {
		t.Errorf("unexpected response text: %s", out.ResponseText)
	}

	recent, _ := mem.GetRecentMessages(ctx, 42, 10)
	if len(recent) != 2 {
		t.Fatalf("expected 2 messages in memory, got %d", len(recent))
	}
	if recent[0].Content != "кофе 150" || recent[1].Content != "Записал: кофе 150 ₽ (еда)" {
		t.Errorf("memory content mismatch")
	}
}
