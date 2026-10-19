package mapper

import (
	"backend/internal/adapter/out/memory/dto"
	"backend/internal/domain/port"
)

func ToDomainMessage(item dto.MemoryMessageItem) port.ChatMessage {
	return port.ChatMessage{
		Role:      item.Role,
		Content:   item.Content,
		CreatedAt: item.CreatedAt,
	}
}

func ToDomainMessages(items []dto.MemoryMessageItem) []port.ChatMessage {
	res := make([]port.ChatMessage, len(items))
	for i, item := range items {
		res[i] = ToDomainMessage(item)
	}
	return res
}

func ToMemoryItem(msg port.ChatMessage) dto.MemoryMessageItem {
	return dto.MemoryMessageItem{
		Role:      msg.Role,
		Content:   msg.Content,
		CreatedAt: msg.CreatedAt,
	}
}
