package llm

import (
	"backend/internal/domain/port"

	"github.com/sashabaranov/go-openai"
)

func ToOpenAIMessages(systemPrompt string, history []port.ChatMessage, userMessage string) []openai.ChatCompletionMessage {
	messages := make([]openai.ChatCompletionMessage, 0, len(history)+2)
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: systemPrompt,
	})

	for _, h := range history {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    h.Role,
			Content: h.Content,
		})
	}

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: userMessage,
	})

	return messages
}

func DefaultOpenAITools() []openai.Tool {
	return []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "add_expense",
				Description: "Сохранить новую трату пользователя",
				Parameters: ToolParameters{
					Type: "object",
					Properties: map[string]ToolParameterProperty{
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
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "list_expenses",
				Description: "Получить список последних трат пользователя или посчитать сумму за период",
				Parameters: ToolParameters{
					Type: "object",
					Properties: map[string]ToolParameterProperty{
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
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "update_expense",
				Description: "Изменить существующую запись траты пользователя по её ID",
				Parameters: ToolParameters{
					Type: "object",
					Properties: map[string]ToolParameterProperty{
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
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "delete_expense",
				Description: "Удалить трату пользователя по её ID",
				Parameters: ToolParameters{
					Type: "object",
					Properties: map[string]ToolParameterProperty{
						"id": {
							Type:        "string",
							Description: "UUID идентификатор удаляемой траты",
						},
					},
					Required: []string{"id"},
				},
			},
		},
	}
}
