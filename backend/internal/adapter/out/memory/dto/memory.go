package dto

import "time"

type MemoryMessageItem struct {
	Role      string
	Content   string
	CreatedAt time.Time
}

type SaveMessageRequest struct {
	UserID    int64
	Role      string
	Content   string
	CreatedAt time.Time
}

type GetMessagesResponse struct {
	Items []MemoryMessageItem
}
