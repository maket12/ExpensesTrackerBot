package usecases

import (
	"context"
	"fmt"
	"strings"
	"time"

	"backend/internal/app/dto"
	apperrs "backend/internal/app/errs"
	"backend/internal/app/mapper"
	"backend/internal/domain/port"
)

type HandleAgentMessageUC struct {
	llmClient    port.LLMClient
	memory       port.ChatMemory
	toolExecutor ToolExecutor
}

func NewHandleAgentMessageUC(
	llmClient port.LLMClient,
	memory port.ChatMemory,
	toolExecutor ToolExecutor,
) *HandleAgentMessageUC {
	return &HandleAgentMessageUC{
		llmClient:    llmClient,
		memory:       memory,
		toolExecutor: toolExecutor,
	}
}

func (uc *HandleAgentMessageUC) Execute(ctx context.Context, input dto.HandleAgentMessageInput) (dto.HandleAgentMessageOutput, error) {
	now := time.Now().UTC()
	systemPrompt := BuildAgentSystemPrompt(now)

	messages := []port.AgentMessage{
		{
			Role:    "system",
			Content: systemPrompt,
		},
	}

	recent, _ := uc.memory.GetRecentMessages(ctx, input.UserID, 8)
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

	const maxIterations = 5
	for i := 0; i < maxIterations; i++ {
		resp, err := uc.llmClient.GenerateResponse(ctx, messages)
		if err != nil {
			return dto.HandleAgentMessageOutput{}, fmt.Errorf("%w: %v", apperrs.ErrAgentFailed, err)
		}

		if len(resp.ToolCalls) == 0 {
			reply := strings.TrimSpace(resp.Content)
			if reply == "" {
				reply = "Готово!"
			}

			_ = uc.memory.AddMessage(ctx, input.UserID, "user", input.Text)
			_ = uc.memory.AddMessage(ctx, input.UserID, "assistant", reply)

			return mapper.ToHandleAgentMessageOutput(reply), nil
		}

		messages = append(messages, port.AgentMessage{
			Role:      "assistant",
			Content:   resp.Content,
			ToolCalls: resp.ToolCalls,
		})

		for _, tc := range resp.ToolCalls {
			toolResult := uc.toolExecutor.Execute(ctx, input.UserID, tc)
			messages = append(messages, port.AgentMessage{
				Role:       "tool",
				Content:    toolResult,
				ToolCallID: tc.ID,
			})
		}
	}

	return dto.HandleAgentMessageOutput{}, apperrs.ErrAgentMaxIterations
}
