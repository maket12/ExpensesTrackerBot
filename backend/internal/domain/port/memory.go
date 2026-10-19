package port

import (
	"context"
	"time"
)

type ChatMessage struct {
	Role      string
	Content   string
	CreatedAt time.Time
}

type ChatMemory interface {
	AddMessage(ctx context.Context, userID int64, role string, content string) error
	GetRecentMessages(ctx context.Context, userID int64, limit int) ([]ChatMessage, error)
	Clear(ctx context.Context, userID int64) error
}
