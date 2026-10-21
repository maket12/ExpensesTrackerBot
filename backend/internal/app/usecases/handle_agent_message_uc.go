package usecases

import (
	"context"

	"backend/internal/app/dto"
	"backend/internal/app/mapper"
	"backend/internal/domain/port"
)

type HandleAgentMessageUC struct {
	assistant port.AIAssistant
	memory    port.ChatMemory
}

func NewHandleAgentMessageUC(assistant port.AIAssistant, memory port.ChatMemory) *HandleAgentMessageUC {
	return &HandleAgentMessageUC{
		assistant: assistant,
		memory:    memory,
	}
}

func (uc *HandleAgentMessageUC) Execute(ctx context.Context, input dto.HandleAgentMessageInput) (dto.HandleAgentMessageOutput, error) {
	history, _ := uc.memory.GetRecentMessages(ctx, input.UserID, 8)

	reply, err := uc.assistant.GenerateReply(ctx, input.UserID, history, input.Text)
	if err != nil {
		return dto.HandleAgentMessageOutput{}, err
	}

	_ = uc.memory.AddMessage(ctx, input.UserID, "user", input.Text)
	_ = uc.memory.AddMessage(ctx, input.UserID, "assistant", reply)

	return mapper.ToHandleAgentMessageOutput(reply), nil
}
