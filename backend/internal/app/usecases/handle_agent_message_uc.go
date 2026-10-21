package usecases

import (
	"context"

	"backend/internal/app/dto"
	ucerrs "backend/internal/app/errs"
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

func (uc *HandleAgentMessageUC) Execute(ctx context.Context, in dto.HandleAgentMessageInput) (*dto.HandleAgentMessageOutput, error) {
	history, err := uc.memory.GetRecentMessages(ctx, in.UserID, 8)
	if err != nil {
		return nil, ucerrs.Wrap(ucerrs.ErrGetRecentMessagesStore, err)
	}

	reply, err := uc.assistant.GenerateReply(ctx, in.UserID, history, in.Text)
	if err != nil {
		return nil, ucerrs.Wrap(ucerrs.ErrGenerateReplyLLM, err)
	}

	if err := uc.memory.AddMessage(ctx, in.UserID, "user", in.Text); err != nil {
		return nil, ucerrs.Wrap(ucerrs.ErrAdMessageStore, err)
	}
	if err := uc.memory.AddMessage(ctx, in.UserID, "assistant", reply); err != nil {
		return nil, ucerrs.Wrap(ucerrs.ErrAdMessageStore, err)
	}

	return mapper.ToHandleAgentMessageOutput(reply), nil
}
