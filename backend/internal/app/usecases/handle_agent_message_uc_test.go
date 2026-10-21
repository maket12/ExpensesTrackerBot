package usecases

import (
	"context"
	"errors"
	"testing"

	"backend/internal/app/dto"
	apperrs "backend/internal/app/errs"
	"backend/internal/domain/port"
	"backend/internal/domain/port/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandleAgentMessageUC_Execute_Success(t *testing.T) {
	assistant := mocks.NewMockAIAssistant(t)
	memory := mocks.NewMockChatMemory(t)
	uc := NewHandleAgentMessageUC(assistant, memory)

	history := []port.ChatMessage{}
	reply := "Записал: кофе 150 ₽ (еда)"

	memory.EXPECT().
		GetRecentMessages(mock.Anything, int64(42), 8).
		Return(history, nil).
		Once()

	assistant.EXPECT().
		GenerateReply(mock.Anything, int64(42), history, "кофе 150").
		Return(reply, nil).
		Once()

	memory.EXPECT().
		AddMessage(mock.Anything, int64(42), "user", "кофе 150").
		Return(nil).
		Once()

	memory.EXPECT().
		AddMessage(mock.Anything, int64(42), "assistant", reply).
		Return(nil).
		Once()

	out, err := uc.Execute(context.Background(), dto.HandleAgentMessageInput{
		UserID: 42,
		Text:   "кофе 150",
	})

	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, reply, out.ResponseText)
}

func TestHandleAgentMessageUC_Execute_MemoryGetRecentError(t *testing.T) {
	assistant := mocks.NewMockAIAssistant(t)
	memory := mocks.NewMockChatMemory(t)
	uc := NewHandleAgentMessageUC(assistant, memory)

	memory.EXPECT().
		GetRecentMessages(mock.Anything, int64(42), 8).
		Return(nil, errors.New("memory error")).
		Once()

	out, err := uc.Execute(context.Background(), dto.HandleAgentMessageInput{
		UserID: 42,
		Text:   "кофе 150",
	})

	assert.Nil(t, out)
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrs.ErrGetRecentMessagesStore))
}

func TestHandleAgentMessageUC_Execute_AssistantError(t *testing.T) {
	assistant := mocks.NewMockAIAssistant(t)
	memory := mocks.NewMockChatMemory(t)
	uc := NewHandleAgentMessageUC(assistant, memory)

	memory.EXPECT().
		GetRecentMessages(mock.Anything, int64(42), 8).
		Return([]port.ChatMessage{}, nil).
		Once()

	assistant.EXPECT().
		GenerateReply(mock.Anything, int64(42), mock.Anything, "кофе 150").
		Return("", errors.New("ai failure")).
		Once()

	out, err := uc.Execute(context.Background(), dto.HandleAgentMessageInput{
		UserID: 42,
		Text:   "кофе 150",
	})

	assert.Nil(t, out)
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrs.ErrGenerateReplyLLM))
}

func TestHandleAgentMessageUC_Execute_MemoryAddUserError(t *testing.T) {
	assistant := mocks.NewMockAIAssistant(t)
	memory := mocks.NewMockChatMemory(t)
	uc := NewHandleAgentMessageUC(assistant, memory)

	memory.EXPECT().
		GetRecentMessages(mock.Anything, int64(42), 8).
		Return([]port.ChatMessage{}, nil).
		Once()

	assistant.EXPECT().
		GenerateReply(mock.Anything, int64(42), mock.Anything, "кофе 150").
		Return("Записал", nil).
		Once()

	memory.EXPECT().
		AddMessage(mock.Anything, int64(42), "user", "кофе 150").
		Return(errors.New("store error")).
		Once()

	out, err := uc.Execute(context.Background(), dto.HandleAgentMessageInput{
		UserID: 42,
		Text:   "кофе 150",
	})

	assert.Nil(t, out)
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrs.ErrAddMessageStore))
}

func TestHandleAgentMessageUC_Execute_MemoryAddAssistantError(t *testing.T) {
	assistant := mocks.NewMockAIAssistant(t)
	memory := mocks.NewMockChatMemory(t)
	uc := NewHandleAgentMessageUC(assistant, memory)

	memory.EXPECT().
		GetRecentMessages(mock.Anything, int64(42), 8).
		Return([]port.ChatMessage{}, nil).
		Once()

	assistant.EXPECT().
		GenerateReply(mock.Anything, int64(42), mock.Anything, "кофе 150").
		Return("Записал", nil).
		Once()

	memory.EXPECT().
		AddMessage(mock.Anything, int64(42), "user", "кофе 150").
		Return(nil).
		Once()

	memory.EXPECT().
		AddMessage(mock.Anything, int64(42), "assistant", "Записал").
		Return(errors.New("store error")).
		Once()

	out, err := uc.Execute(context.Background(), dto.HandleAgentMessageInput{
		UserID: 42,
		Text:   "кофе 150",
	})

	assert.Nil(t, out)
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperrs.ErrAddMessageStore))
}
