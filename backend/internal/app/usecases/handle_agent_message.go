package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"backend/internal/app/dto"
	apperrs "backend/internal/app/errs"
	"backend/internal/domain/port"
	"backend/pkg/utils"
)

type AgentUseCase struct {
	expenseUseCase *ExpenseUseCase
	llmClient      port.LLMClient
	memory         port.ChatMemory
}

func NewAgentUseCase(
	expenseUseCase *ExpenseUseCase,
	llmClient port.LLMClient,
	memory port.ChatMemory,
) *AgentUseCase {
	return &AgentUseCase{
		expenseUseCase: expenseUseCase,
		llmClient:      llmClient,
		memory:         memory,
	}
}

type addExpenseArgs struct {
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
	SpentAt  string  `json:"spent_at"`
	Comment  string  `json:"comment"`
}

type listExpensesArgs struct {
	Limit    int     `json:"limit"`
	Category *string `json:"category"`
	From     *string `json:"from"`
	To       *string `json:"to"`
}

type updateExpenseArgs struct {
	ID       string   `json:"id"`
	Amount   *float64 `json:"amount"`
	Category *string  `json:"category"`
	SpentAt  *string  `json:"spent_at"`
	Comment  *string  `json:"comment"`
}

type deleteExpenseArgs struct {
	ID string `json:"id"`
}

func (a *AgentUseCase) getTools() []port.AgentTool {
	return []port.AgentTool{
		{
			Name:        "add_expense",
			Description: "Сохранить новую трату пользователя",
			Parameters: port.ToolParameters{
				Type: "object",
				Properties: map[string]port.ToolParameterProperty{
					"amount": {
						Type:        "number",
						Description: "Сумма траты больше 0",
					},
					"category": {
						Type:        "string",
						Description: "Категория траты строго из списка",
						Enum: []string{
							"еда", "транспорт", "жильё", "развлечения", "здоровье", "покупки", "другое",
						},
					},
					"spent_at": {
						Type:        "string",
						Description: "Дата и время траты в формате ISO 8601 (RFC3339)",
					},
					"comment": {
						Type:        "string",
						Description: "Необязательный комментарий или описание (например: кофе, такси)",
					},
				},
				Required: []string{"amount", "category", "spent_at"},
			},
		},
		{
			Name:        "list_expenses",
			Description: "Получить список последних трат пользователя или посчитать сумму за период",
			Parameters: port.ToolParameters{
				Type: "object",
				Properties: map[string]port.ToolParameterProperty{
					"limit": {
						Type:        "integer",
						Description: "Максимальное число записей (по умолчанию 10)",
					},
					"category": {
						Type:        "string",
						Description: "Фильтр по категории",
						Enum: []string{
							"еда", "транспорт", "жильё", "развлечения", "здоровье", "покупки", "другое",
						},
					},
					"from": {
						Type:        "string",
						Description: "Начало периода в формате RFC3339",
					},
					"to": {
						Type:        "string",
						Description: "Конец периода в формате RFC3339",
					},
				},
			},
		},
		{
			Name:        "update_expense",
			Description: "Изменить существующую запись траты пользователя по её ID",
			Parameters: port.ToolParameters{
				Type: "object",
				Properties: map[string]port.ToolParameterProperty{
					"id": {
						Type:        "string",
						Description: "UUID идентификатор траты",
					},
					"amount": {
						Type:        "number",
						Description: "Новая сумма",
					},
					"category": {
						Type:        "string",
						Description: "Новая категория",
						Enum: []string{
							"еда", "транспорт", "жильё", "развлечения", "здоровье", "покупки", "другое",
						},
					},
					"spent_at": {
						Type:        "string",
						Description: "Новая дата и время в формате RFC3339",
					},
					"comment": {
						Type:        "string",
						Description: "Новый комментарий",
					},
				},
				Required: []string{"id"},
			},
		},
		{
			Name:        "delete_expense",
			Description: "Удалить трату пользователя по её ID",
			Parameters: port.ToolParameters{
				Type: "object",
				Properties: map[string]port.ToolParameterProperty{
					"id": {
						Type:        "string",
						Description: "UUID идентификатор удаляемой траты",
					},
				},
				Required: []string{"id"},
			},
		},
	}
}

func (a *AgentUseCase) buildSystemPrompt(now time.Time) string {
	weekdays := map[time.Weekday]string{
		time.Monday:    "Понедельник",
		time.Tuesday:   "Вторник",
		time.Wednesday: "Среда",
		time.Thursday:  "Четверг",
		time.Friday:    "Пятница",
		time.Saturday:  "Суббота",
		time.Sunday:    "Воскресенье",
	}

	return fmt.Sprintf(`Ты — персональный финансовый ассистент по учету семейных расходов.
Твоя задача — понимать сообщения пользователя на естественном русском языке и управлять его личными тратами через вызовы функций.

Текущий момент времени: %s, %s (%s).
Разрешенные категории: еда, транспорт, жильё, развлечения, здоровье, покупки, другое.

Правила работы:
1. Вызывай соответствующие функции (tool calls) для любых операций:
   - добавить трату: add_expense
   - посмотреть / посчитать траты: list_expenses
   - изменить существующую трату: update_expense
   - удалить трату: delete_expense
2. Если пользователь просит изменить («последнее было не 350, а 450») или удалить («удали такси») трату без указания UUID:
   - Сначала вызови list_expenses с limit=5, чтобы найти нужную трату по контексту или взять самую последнюю.
   - Затем вызови update_expense или delete_expense с найденным id.
3. Относительные даты разрешай строго относительно текущего момента:
   - «сегодня» -> дата текущего дня.
   - «вчера» -> текущий день минус 1 день.
   - «в этом месяце» -> с 1 числа текущего месяца по текущий момент.
   - «за неделю» -> последние 7 дней.
4. Отвечай кратко, дружелюбно и вежливо на русском языке, подтверждая выполненное действие (например: «Записал: кофе 350 ₽ (еда)»).`,
		weekdays[now.Weekday()],
		now.Format("02.01.2006 15:04:05"),
		now.Format(time.RFC3339),
	)
}

func (a *AgentUseCase) HandleMessage(ctx context.Context, input dto.HandleAgentMessageInput) (dto.HandleAgentMessageOutput, error) {
	now := time.Now().UTC()
	systemPrompt := a.buildSystemPrompt(now)

	messages := []port.AgentMessage{
		{
			Role:    "system",
			Content: systemPrompt,
		},
	}

	recent, _ := a.memory.GetRecentMessages(ctx, input.UserID, 8)
	for _, m := range recent {
		messages = append(messages, port.AgentMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	messages = append(messages, port.AgentMessage{
		Role:    "user",
		Content: input.Text,
	})

	tools := a.getTools()

	maxIterations := 5
	for i := 0; i < maxIterations; i++ {
		resp, err := a.llmClient.GenerateResponse(ctx, messages, tools)
		if err != nil {
			return dto.HandleAgentMessageOutput{}, fmt.Errorf("%w: %v", apperrs.ErrAgentFailed, err)
		}

		if len(resp.ToolCalls) == 0 {
			reply := strings.TrimSpace(resp.Content)
			if reply == "" {
				reply = "Готово!"
			}

			_ = a.memory.AddMessage(ctx, input.UserID, "user", input.Text)
			_ = a.memory.AddMessage(ctx, input.UserID, "assistant", reply)

			return dto.HandleAgentMessageOutput{
				ResponseText: reply,
			}, nil
		}

		messages = append(messages, port.AgentMessage{
			Role:      "assistant",
			Content:   resp.Content,
			ToolCalls: resp.ToolCalls,
		})

		for _, tc := range resp.ToolCalls {
			toolResult := a.executeTool(ctx, input.UserID, tc)
			messages = append(messages, port.AgentMessage{
				Role:       "tool",
				Content:    toolResult,
				ToolCallID: tc.ID,
			})
		}
	}

	return dto.HandleAgentMessageOutput{}, apperrs.ErrAgentMaxIterations
}

func (a *AgentUseCase) executeTool(ctx context.Context, userID int64, tc port.ToolCall) string {
	switch tc.Name {
	case "add_expense":
		var args addExpenseArgs
		if err := json.Unmarshal([]byte(tc.Arguments), &args); err != nil {
			return fmt.Sprintf(`{"error":"некорректные параметры: %s"}`, err.Error())
		}

		spentAt, err := time.Parse(time.RFC3339, args.SpentAt)
		if err != nil {
			spentAt = time.Now().UTC()
		}

		out, err := a.expenseUseCase.CreateExpense(ctx, dto.CreateExpenseInput{
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
		var args listExpensesArgs
		_ = json.Unmarshal([]byte(tc.Arguments), &args)

		limit := args.Limit
		if limit <= 0 {
			limit = 10
		}

		var fromTime, toTime *time.Time
		if args.From != nil {
			if t, err := time.Parse(time.RFC3339, *args.From); err == nil {
				fromTime = utils.VPtr(t)
			}
		}
		if args.To != nil {
			if t, err := time.Parse(time.RFC3339, *args.To); err == nil {
				toTime = utils.VPtr(t)
			}
		}

		out, err := a.expenseUseCase.ListExpenses(ctx, dto.ListExpensesInput{
			UserID:   utils.VPtr(userID),
			Category: args.Category,
			From:     fromTime,
			To:       toTime,
			Limit:    limit,
		})
		if err != nil {
			return fmt.Sprintf(`{"error":"ошибка получения списка трат: %s"}`, err.Error())
		}

		var total float64
		for _, e := range out.Expenses {
			total += e.Amount
		}

		res := map[string]any{
			"expenses":    out.Expenses,
			"count":       len(out.Expenses),
			"total_sum":   total,
		}
		data, _ := json.Marshal(res)
		return string(data)

	case "update_expense":
		var args updateExpenseArgs
		if err := json.Unmarshal([]byte(tc.Arguments), &args); err != nil {
			return fmt.Sprintf(`{"error":"некорректные параметры: %s"}`, err.Error())
		}

		existing, err := a.expenseUseCase.GetExpense(ctx, dto.GetExpenseInput{ID: args.ID})
		if err != nil {
			return fmt.Sprintf(`{"error":"трата с id %s не найдена"}`, args.ID)
		}

		amount := existing.Expense.Amount
		if args.Amount != nil {
			amount = *args.Amount
		}

		category := existing.Expense.Category
		if args.Category != nil {
			category = *args.Category
		}

		spentAt := existing.Expense.SpentAt
		if args.SpentAt != nil {
			if t, err := time.Parse(time.RFC3339, *args.SpentAt); err == nil {
				spentAt = t
			}
		}

		comment := existing.Expense.Comment
		if args.Comment != nil {
			comment = *args.Comment
		}

		out, err := a.expenseUseCase.UpdateExpense(ctx, dto.UpdateExpenseInput{
			ID:            args.ID,
			UserID:        userID,
			Amount:        amount,
			Category:      category,
			SpentAt:       spentAt,
			Comment:       comment,
			EnforceUserID: true,
		})
		if err != nil {
			return fmt.Sprintf(`{"error":"ошибка обновления: %s"}`, err.Error())
		}
		data, _ := json.Marshal(out.Expense)
		return string(data)

	case "delete_expense":
		var args deleteExpenseArgs
		if err := json.Unmarshal([]byte(tc.Arguments), &args); err != nil {
			return fmt.Sprintf(`{"error":"некорректные параметры: %s"}`, err.Error())
		}

		out, err := a.expenseUseCase.DeleteExpense(ctx, dto.DeleteExpenseInput{
			ID:            args.ID,
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
