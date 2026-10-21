package usecases

import (
	"context"
	"errors"
	"testing"

	"backend/internal/adapter/out/memory"
	"backend/internal/app/dto"
	"backend/internal/domain/port"
)

type mockAIAssistant struct {
	reply string
	err   error
}

func (m *mockAIAssistant) GenerateReply(_ context.Context, _ int64, _ []port.ChatMessage, _ string) (string, error) {
	return m.reply, m.err
}

func TestHandleAgentMessageUC_Execute(t *testing.T) {
	ctx := context.Background()
	mem := memory.NewStore()
	assistant := &mockAIAssistant{reply: "Записал: кофе 150 ₽ (еда)"}

	uc := NewHandleAgentMessageUC(assistant, mem)

	out, err := uc.Execute(ctx, dto.HandleAgentMessageInput{
		UserID: 42,
		Text:   "кофе 150",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.ResponseText != "Записал: кофе 150 ₽ (еда)" {
		t.Errorf("unexpected response text: %s", out.ResponseText)
	}

	recent, _ := mem.GetRecentMessages(ctx, 42, 10)
	if len(recent) != 2 {
		t.Fatalf("expected 2 messages in memory, got %d", len(recent))
	}
	if recent[0].Content != "кофе 150" || recent[1].Content != "Записал: кофе 150 ₽ (еда)" {
		t.Errorf("memory content mismatch")
	}
}

func TestHandleAgentMessageUC_Error(t *testing.T) {
	ctx := context.Background()
	mem := memory.NewStore()
	expectedErr := errors.New("ai failure")
	assistant := &mockAIAssistant{err: expectedErr}

	uc := NewHandleAgentMessageUC(assistant, mem)

	_, err := uc.Execute(ctx, dto.HandleAgentMessageInput{
		UserID: 42,
		Text:   "кофе 150",
	})
	if err == nil || !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}
