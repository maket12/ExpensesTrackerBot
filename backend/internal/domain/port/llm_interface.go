package port

import "context"

type AIAssistant interface {
	GenerateReply(ctx context.Context, userID int64, history []ChatMessage, userMessage string) (string, error)
}
