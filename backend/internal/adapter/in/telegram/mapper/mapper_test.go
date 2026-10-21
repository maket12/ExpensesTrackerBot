package mapper

import (
	"testing"

	tgdto "backend/internal/adapter/in/telegram/dto"
	appdto "backend/internal/app/dto"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
)

func TestToAgentInput(t *testing.T) {
	req := tgdto.ProcessMessageRequest{
		UserID: int64(gofakeit.Number(1000, 999999)),
		Text:   gofakeit.Sentence(3),
	}

	input := ToAgentInput(req)
	assert.Equal(t, req.UserID, input.UserID)
	assert.Equal(t, req.Text, input.Text)
}

func TestToTelegramResponse(t *testing.T) {
	t.Run("non-nil output mapped correctly", func(t *testing.T) {
		reply := gofakeit.Sentence(4)
		out := &appdto.HandleAgentMessageOutput{
			ResponseText: reply,
		}

		resp := ToTelegramResponse(out)
		assert.Equal(t, reply, resp.Text)
	})

	t.Run("nil output returns empty response", func(t *testing.T) {
		resp := ToTelegramResponse(nil)
		assert.Equal(t, "", resp.Text)
	})
}
