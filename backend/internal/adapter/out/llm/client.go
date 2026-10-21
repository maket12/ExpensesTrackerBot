package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	appdto "backend/internal/app/dto"
	apperrs "backend/internal/app/errs"
	"backend/internal/app/usecases"
	"backend/internal/domain/port"
	pkgerrs "backend/pkg/errs"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

type chatCompleter interface {
	CreateChatCompletion(ctx context.Context, request openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
}

type Client struct {
	client   chatCompleter
	model    string
	tools    []openai.Tool
	createUC *usecases.CreateExpenseUC
	listUC   *usecases.ListExpensesUC
	updateUC *usecases.UpdateExpenseUC
	deleteUC *usecases.DeleteExpenseUC
}

func NewClient(
	baseURL, apiKey, model string,
	createUC *usecases.CreateExpenseUC,
	listUC *usecases.ListExpensesUC,
	updateUC *usecases.UpdateExpenseUC,
	deleteUC *usecases.DeleteExpenseUC,
) (*Client, error) {
	if baseURL == "" {
		return nil, pkgerrs.NewValueRequiredError("llm_base_url")
	}
	if apiKey == "" {
		return nil, pkgerrs.NewValueRequiredError("llm_api_key")
	}
	if model == "" {
		return nil, pkgerrs.NewValueRequiredError("llm_model")
	}

	cfg := openai.DefaultConfig(apiKey)
	cfg.BaseURL = baseURL

	return &Client{
		client:   openai.NewClientWithConfig(cfg),
		model:    model,
		tools:    DefaultOpenAITools(),
		createUC: createUC,
		listUC:   listUC,
		updateUC: updateUC,
		deleteUC: deleteUC,
	}, nil
}

func (c *Client) GenerateReply(ctx context.Context, userID int64, history []port.ChatMessage, userMessage string) (string, error) {
	now := time.Now().UTC()
	systemPrompt := BuildSystemPrompt(now)
	messages := ToOpenAIMessages(systemPrompt, history, userMessage)

	const maxIterations = 5
	for i := 0; i < maxIterations; i++ {
		req := openai.ChatCompletionRequest{
			Model:       c.model,
			Messages:    messages,
			Temperature: 0.1,
			Tools:       c.tools,
			ToolChoice:  "auto",
		}

		resp, err := c.client.CreateChatCompletion(ctx, req)
		if err != nil {
			return "", fmt.Errorf("%w: %v", apperrs.ErrAgentFailed, err)
		}

		if len(resp.Choices) == 0 {
			return "", errors.New("empty response from llm")
		}

		choice := resp.Choices[0]
		msg := choice.Message

		if len(msg.ToolCalls) == 0 {
			reply := strings.TrimSpace(msg.Content)
			if reply == "" {
				reply = "Готово!"
			}
			return reply, nil
		}

		messages = append(messages, msg)

		for _, tc := range msg.ToolCalls {
			toolResult := c.executeTool(ctx, userID, tc)
			messages = append(messages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    toolResult,
				ToolCallID: tc.ID,
			})
		}
	}

	return "", apperrs.ErrAgentMaxIterations
}

func (c *Client) executeTool(ctx context.Context, userID int64, tc openai.ToolCall) string {
	switch tc.Function.Name {
	case "add_expense":
		var args AddExpenseToolPayload
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
			return fmt.Sprintf(`{"error":"некорректные параметры: %s"}`, err.Error())
		}

		spentAt, err := time.Parse(time.RFC3339, args.SpentAt)
		if err != nil {
			spentAt = time.Now().UTC()
		}

		out, err := c.createUC.Execute(ctx, appdto.CreateExpenseInput{
			UserID:   userID,
			Amount:   args.Amount,
			Category: args.Category,
			SpentAt:  spentAt,
			Comment:  args.Comment,
		})
		if err != nil {
			return fmt.Sprintf(`{"error":"ошибка создания траты: %s"}`, err.Error())
		}
		data, _ := json.Marshal(out.Expense)
		return string(data)

	case "list_expenses":
		var args ListExpensesToolPayload
		_ = json.Unmarshal([]byte(tc.Function.Arguments), &args)

		limit := args.Limit
		if limit <= 0 {
			limit = 10
		}

		var fromTime, toTime *time.Time
		if args.From != nil {
			if t, err := time.Parse(time.RFC3339, *args.From); err == nil {
				fromTime = &t
			}
		}
		if args.To != nil {
			if t, err := time.Parse(time.RFC3339, *args.To); err == nil {
				toTime = &t
			}
		}

		out, err := c.listUC.Execute(ctx, appdto.ListExpensesInput{
			UserID:   &userID,
			Category: args.Category,
			From:     fromTime,
			To:       toTime,
			Limit:    limit,
		})
		if err != nil {
			return fmt.Sprintf(`{"error":"ошибка получения списка трат: %s"}`, err.Error())
		}

		var total float64
		for _, item := range out.Expenses {
			total += item.Amount
		}

		res := map[string]any{
			"expenses":  out.Expenses,
			"count":     len(out.Expenses),
			"total_sum": total,
		}
		data, _ := json.Marshal(res)
		return string(data)

	case "update_expense":
		var args UpdateExpenseToolPayload
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
			return fmt.Sprintf(`{"error":"некорректные параметры: %s"}`, err.Error())
		}

		var spentAt *time.Time
		if args.SpentAt != nil {
			t, err := time.Parse(time.RFC3339, *args.SpentAt)
			if err != nil {
				return fmt.Sprintf(`{"error":"некорректный формат spent_at: %s"}`, err.Error())
			}
			spentAt = &t
		}

		expenseID, err := uuid.Parse(args.ID)
		if err != nil {
			return fmt.Sprintf(`{"error":"некорректный формат id: %s"}`, err.Error())
		}

		out, err := c.updateUC.Execute(ctx, appdto.UpdateExpenseInput{
			ID:            expenseID,
			UserID:        userID,
			Amount:        args.Amount,
			Category:      args.Category,
			SpentAt:       spentAt,
			Comment:       args.Comment,
			EnforceUserID: true,
		})
		if err != nil {
			return fmt.Sprintf(`{"error":"ошибка обновления: %s"}`, err.Error())
		}
		data, _ := json.Marshal(out.Expense)
		return string(data)

	case "delete_expense":
		var args DeleteExpenseToolPayload
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
			return fmt.Sprintf(`{"error":"некорректные параметры: %s"}`, err.Error())
		}

		expenseID, err := uuid.Parse(args.ID)
		if err != nil {
			return fmt.Sprintf(`{"error":"некорректный формат id: %s"}`, err.Error())
		}

		out, err := c.deleteUC.Execute(ctx, appdto.DeleteExpenseInput{
			ID:            expenseID,
			UserID:        userID,
			EnforceUserID: true,
		})
		if err != nil {
			return fmt.Sprintf(`{"error":"ошибка удаления: %s"}`, err.Error())
		}
		data, _ := json.Marshal(out)
		return string(data)

	default:
		return `{"error":"неизвестный инструмент"}`
	}
}
