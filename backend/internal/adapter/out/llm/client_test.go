package llm

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"backend/internal/app/usecases"
	"backend/internal/domain/models"
	"backend/internal/domain/port"
	pkgerrs "backend/pkg/errs"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

type mockRepo struct {
	expenses map[uuid.UUID]*models.Expense
}

func newMockRepo() *mockRepo {
	return &mockRepo{expenses: make(map[uuid.UUID]*models.Expense)}
}

func (m *mockRepo) Create(_ context.Context, expense *models.Expense) (*models.Expense, error) {
	m.expenses[expense.ID()] = expense
	return expense, nil
}

func (m *mockRepo) Get(_ context.Context, id uuid.UUID) (*models.Expense, error) {
	e, ok := m.expenses[id]
	if !ok {
		return nil, pkgerrs.NewObjectNotFoundError("expense", id)
	}
	return e, nil
}

func (m *mockRepo) Update(_ context.Context, expense *models.Expense) (*models.Expense, error) {
	if _, ok := m.expenses[expense.ID()]; !ok {
		return nil, pkgerrs.NewObjectNotFoundError("expense", expense.ID())
	}
	m.expenses[expense.ID()] = expense
	return expense, nil
}

func (m *mockRepo) Delete(_ context.Context, id uuid.UUID) error {
	if _, ok := m.expenses[id]; !ok {
		return pkgerrs.NewObjectNotFoundError("expense", id)
	}
	delete(m.expenses, id)
	return nil
}

func (m *mockRepo) List(_ context.Context, filter port.ExpenseFilter) ([]*models.Expense, error) {
	var res []*models.Expense
	for _, e := range m.expenses {
		if filter.UserID != nil && e.UserID() != *filter.UserID {
			continue
		}
		res = append(res, e)
	}
	return res, nil
}

func (m *mockRepo) ListLatest(_ context.Context, userID int64, limit int) ([]*models.Expense, error) {
	return nil, nil
}

func (m *mockRepo) ListUserIDS(_ context.Context) ([]int64, error) {
	return nil, nil
}

type mockCompleter struct {
	responses []openai.ChatCompletionResponse
	callIndex int
}

func (m *mockCompleter) CreateChatCompletion(_ context.Context, _ openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	if m.callIndex < len(m.responses) {
		r := m.responses[m.callIndex]
		m.callIndex++
		return r, nil
	}
	return openai.ChatCompletionResponse{
		Choices: []openai.ChatCompletionChoice{
			{Message: openai.ChatCompletionMessage{Content: "Готово!"}},
		},
	}, nil
}

func TestBuildSystemPrompt(t *testing.T) {
	now := time.Date(2026, 10, 21, 14, 0, 0, 0, time.UTC)
	prompt := BuildSystemPrompt(now)
	if !strings.Contains(prompt, "Среда") {
		t.Errorf("expected prompt to contain 'Среда'")
	}
	if !strings.Contains(prompt, "21.10.2026") {
		t.Errorf("expected prompt to contain date '21.10.2026'")
	}
}

func TestClient_GenerateReply_Direct(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepo()
	completer := &mockCompleter{
		responses: []openai.ChatCompletionResponse{
			{
				Choices: []openai.ChatCompletionChoice{
					{Message: openai.ChatCompletionMessage{Content: "Привет! Чем помочь?"}},
				},
			},
		},
	}

	c := &Client{
		client:   completer,
		model:    "gpt-4o-mini",
		tools:    DefaultOpenAITools(),
		createUC: usecases.NewCreateExpenseUC(repo),
		listUC:   usecases.NewListExpensesUC(repo),
		updateUC: usecases.NewUpdateExpenseUC(repo),
		deleteUC: usecases.NewDeleteExpenseUC(repo),
	}

	reply, err := c.GenerateReply(ctx, 100, nil, "привет")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reply != "Привет! Чем помочь?" {
		t.Errorf("unexpected reply: %s", reply)
	}
}

func TestClient_GenerateReply_WithToolCall(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepo()

	addArgs, _ := json.Marshal(map[string]any{
		"amount":   350.0,
		"category": "еда",
		"spent_at": time.Now().UTC().Add(-time.Hour).Format(time.RFC3339),
		"comment":  "кофе",
	})

	completer := &mockCompleter{
		responses: []openai.ChatCompletionResponse{
			{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							ToolCalls: []openai.ToolCall{
								{
									ID:   "call-1",
									Type: openai.ToolTypeFunction,
									Function: openai.FunctionCall{
										Name:      "add_expense",
										Arguments: string(addArgs),
									},
								},
							},
						},
					},
				},
			},
			{
				Choices: []openai.ChatCompletionChoice{
					{Message: openai.ChatCompletionMessage{Content: "Записал: кофе 350 ₽ (еда)"}},
				},
			},
		},
	}

	c := &Client{
		client:   completer,
		model:    "gpt-4o-mini",
		tools:    DefaultOpenAITools(),
		createUC: usecases.NewCreateExpenseUC(repo),
		listUC:   usecases.NewListExpensesUC(repo),
		updateUC: usecases.NewUpdateExpenseUC(repo),
		deleteUC: usecases.NewDeleteExpenseUC(repo),
	}

	reply, err := c.GenerateReply(ctx, 100, nil, "кофе 350")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reply != "Записал: кофе 350 ₽ (еда)" {
		t.Errorf("unexpected reply: %s", reply)
	}

	if len(repo.expenses) != 1 {
		t.Fatalf("expected 1 expense created in repo, got %d", len(repo.expenses))
	}
}

func TestClient_ExecuteTool_AllTools(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepo()
	c := &Client{
		createUC: usecases.NewCreateExpenseUC(repo),
		listUC:   usecases.NewListExpensesUC(repo),
		updateUC: usecases.NewUpdateExpenseUC(repo),
		deleteUC: usecases.NewDeleteExpenseUC(repo),
	}

	userID := int64(100)

	addArgs, _ := json.Marshal(map[string]any{
		"amount":   500.0,
		"category": "транспорт",
		"spent_at": time.Now().UTC().Add(-time.Hour).Format(time.RFC3339),
		"comment":  "такси",
	})
	resAdd := c.executeTool(ctx, userID, openai.ToolCall{
		ID: "tc-1",
		Function: openai.FunctionCall{
			Name:      "add_expense",
			Arguments: string(addArgs),
		},
	})
	if strings.Contains(resAdd, "error") {
		t.Fatalf("unexpected add_expense error: %s", resAdd)
	}

	var added map[string]any
	_ = json.Unmarshal([]byte(resAdd), &added)
	expenseID := added["ID"].(string)

	listArgs, _ := json.Marshal(map[string]any{"limit": 5})
	resList := c.executeTool(ctx, userID, openai.ToolCall{
		ID: "tc-2",
		Function: openai.FunctionCall{
			Name:      "list_expenses",
			Arguments: string(listArgs),
		},
	})
	if strings.Contains(resList, "error") || !strings.Contains(resList, "500") {
		t.Errorf("unexpected list response: %s", resList)
	}

	updateArgs, _ := json.Marshal(map[string]any{
		"id":     expenseID,
		"amount": 600.0,
	})
	resUpdate := c.executeTool(ctx, userID, openai.ToolCall{
		ID: "tc-3",
		Function: openai.FunctionCall{
			Name:      "update_expense",
			Arguments: string(updateArgs),
		},
	})
	if strings.Contains(resUpdate, "error") {
		t.Errorf("unexpected update response: %s", resUpdate)
	}

	deleteArgs, _ := json.Marshal(map[string]any{
		"id": expenseID,
	})
	resDelete := c.executeTool(ctx, userID, openai.ToolCall{
		ID: "tc-4",
		Function: openai.FunctionCall{
			Name:      "delete_expense",
			Arguments: string(deleteArgs),
		},
	})
	if strings.Contains(resDelete, "error") {
		t.Errorf("unexpected delete response: %s", resDelete)
	}

	unknownRes := c.executeTool(ctx, userID, openai.ToolCall{
		ID: "tc-5",
		Function: openai.FunctionCall{
			Name: "invalid_tool",
		},
	})
	if !strings.Contains(unknownRes, "error") {
		t.Errorf("expected error for unknown tool")
	}
}
