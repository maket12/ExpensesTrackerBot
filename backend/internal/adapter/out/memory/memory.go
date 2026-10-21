package memory

import (
	"context"
	"sync"
	"time"

	"backend/internal/domain/port"
)

type Store struct {
	mu      sync.RWMutex
	history map[int64][]port.ChatMessage
}

func NewStore() *Store {
	return &Store{
		history: make(map[int64][]port.ChatMessage),
	}
}

func (s *Store) AddMessage(_ context.Context, userID int64, role string, content string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.history[userID] = append(s.history[userID], port.ChatMessage{
		Role:      role,
		Content:   content,
		CreatedAt: time.Now().UTC(),
	})

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
	selected := make([]port.ChatMessage, limit)
	copy(selected, items[start:])

	return selected, nil
}

func (s *Store) Clear(_ context.Context, userID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.history, userID)
	return nil
}
