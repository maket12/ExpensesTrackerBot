package mapper

import (
	"backend/internal/domain/port"

	"github.com/sashabaranov/go-openai"
)

func ToOpenAIMessages(messages []port.AgentMessage) []openai.ChatCompletionMessage {
	result := make([]openai.ChatCompletionMessage, len(messages))
	for i, m := range messages {
		msg := openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		}
		if m.ToolCallID != "" {
			msg.ToolCallID = m.ToolCallID
		}
		if len(m.ToolCalls) > 0 {
			tcs := make([]openai.ToolCall, len(m.ToolCalls))
			for j, tc := range m.ToolCalls {
				tcs[j] = openai.ToolCall{
					ID:   tc.ID,
					Type: openai.ToolTypeFunction,
					Function: openai.FunctionCall{
						Name:      tc.Name,
						Arguments: tc.Arguments,
					},
				}
			}
			msg.ToolCalls = tcs
		}
		result[i] = msg
	}
	return result
}

func ToOpenAITools(tools []port.AgentTool) []openai.Tool {
	result := make([]openai.Tool, len(tools))
	for i, t := range tools {
		result[i] = openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			},
		}
	}
	return result
}

func ToDomainAgentResponse(choice openai.ChatCompletionChoice) port.AgentResponse {
	var toolCalls []port.ToolCall
	for _, tc := range choice.Message.ToolCalls {
		toolCalls = append(toolCalls, port.ToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: tc.Function.Arguments,
		})
	}

	return port.AgentResponse{
		Content:   choice.Message.Content,
		ToolCalls: toolCalls,
	}
}
