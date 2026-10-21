package llm

import (
	"testing"
	"time"

	"backend/internal/domain/port"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToOpenAIMessages(t *testing.T) {
	t.Run("empty history", func(t *testing.T) {
		systemPrompt := gofakeit.Sentence(5)
		userMessage := gofakeit.Sentence(3)

		msgs := ToOpenAIMessages(systemPrompt, nil, userMessage)
		require.Len(t, msgs, 2)
		assert.Equal(t, openai.ChatMessageRoleSystem, msgs[0].Role)
		assert.Equal(t, systemPrompt, msgs[0].Content)
		assert.Equal(t, openai.ChatMessageRoleUser, msgs[1].Role)
		assert.Equal(t, userMessage, msgs[1].Content)
	})

	t.Run("populated history with fake messages", func(t *testing.T) {
		systemPrompt := gofakeit.Sentence(5)
		userMessage := gofakeit.Sentence(3)

		count := gofakeit.Number(2, 5)
		history := make([]port.ChatMessage, count)
		for i := 0; i < count; i++ {
			role := "user"
			if i%2 == 1 {
				role = "assistant"
			}
			history[i] = port.ChatMessage{
				Role:      role,
				Content:   gofakeit.Sentence(4),
				CreatedAt: time.Now().UTC().Add(-time.Duration(count-i) * time.Minute),
			}
		}

		msgs := ToOpenAIMessages(systemPrompt, history, userMessage)
		require.Len(t, msgs, count+2)

		assert.Equal(t, openai.ChatMessageRoleSystem, msgs[0].Role)
		assert.Equal(t, systemPrompt, msgs[0].Content)

		for i := 0; i < count; i++ {
			assert.Equal(t, history[i].Role, msgs[i+1].Role)
			assert.Equal(t, history[i].Content, msgs[i+1].Content)
		}

		assert.Equal(t, openai.ChatMessageRoleUser, msgs[count+1].Role)
		assert.Equal(t, userMessage, msgs[count+1].Content)
	})
}

func TestDefaultOpenAITools(t *testing.T) {
	tools := DefaultOpenAITools()
	require.Len(t, tools, 4)

	toolNames := make(map[string]openai.Tool)
	for _, tool := range tools {
		assert.Equal(t, openai.ToolTypeFunction, tool.Type)
		require.NotNil(t, tool.Function)
		toolNames[tool.Function.Name] = tool
	}

	assert.Contains(t, toolNames, "add_expense")
	assert.Contains(t, toolNames, "list_expenses")
	assert.Contains(t, toolNames, "update_expense")
	assert.Contains(t, toolNames, "delete_expense")

	addTool := toolNames["add_expense"]
	params, ok := addTool.Function.Parameters.(ToolParameters)
	require.True(t, ok)
	assert.Equal(t, "object", params.Type)
	assert.Contains(t, params.Required, "amount")
	assert.Contains(t, params.Required, "category")
	assert.Contains(t, params.Required, "spent_at")

	delTool := toolNames["delete_expense"]
	delParams, ok := delTool.Function.Parameters.(ToolParameters)
	require.True(t, ok)
	assert.Contains(t, delParams.Required, "id")
}
