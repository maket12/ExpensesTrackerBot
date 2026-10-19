package memory

import (
	"context"
	"strings"
	"sync"
	"time"

	"backend/internal/adapter/out/memory/dto"
	"backend/internal/adapter/out/memory/errs"
	"backend/internal/adapter/out/memory/mapper"
	"backend/internal/domain/port"
)

type Store struct {
	mu       sync.RWMutex
	capacity int
	history  map[int64][]dto.MemoryMessageItem
}

func NewStore(capacity int) *Store {
	if capacity <= 0 {
		capacity = 15
	}
	return &Store{
		capacity: capacity,
		history:  make(map[int64][]dto.MemoryMessageItem),
	}
}

func (s *Store) AddMessage(_ context.Context, userID int64, role string, content string) error {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return errs.ErrEmptyMessage
	}

	req := dto.SaveMessageRequest{
		UserID:    userID,
		Role:      role,
		Content:   trimmed,
		CreatedAt: time.Now().UTC(),
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	items := s.history[req.UserID]
	item := dto.MemoryMessageItem{
		Role:      req.Role,
		Content:   req.Content,
		CreatedAt: req.CreatedAt,
	}

	items = append(items, item)
	if len(items) > s.capacity {
		items = items[len(items)-s.capacity:]
	}
	s.history[req.UserID] = items

	return nil
}

func (s *Store) GetRecentMessages(_ context.Context, userID int64, limit int) ([]port.ChatMessage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items, ok := s.history[userID]
	if !ok || len(items) == 0 {
		return []port.ChatMessage{}, nil
	}

	if limit <= 0 || limit > len(items) {
		limit = len(items)
	}

	start := len(items) - limit
	selected := make([]dto.MemoryMessageItem, limit)
	copy(selected, items[start:])

	resp := dto.GetMessagesResponse{Items: selected}
	return mapper.ToDomainMessages(resp.Items), nil
}

func (s *Store) Clear(_ context.Context, userID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.history, userID)
	return nil
}
