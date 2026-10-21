package memory

import (
	"context"
	"testing"
)

func TestStore_AddAndGetRecentMessages(t *testing.T) {
	ctx := context.Background()
	store := NewStore()

	userID := int64(100)

	msgs, err := store.GetRecentMessages(ctx, userID, 5)
	if err != nil {
		t.Fatalf("unexpected error getting messages for empty user: %v", err)
	}
	if len(msgs) != 0 {
		t.Fatalf("expected 0 messages, got %d", len(msgs))
	}

	for i := 1; i <= 5; i++ {
		err := store.AddMessage(ctx, userID, "user", "msg")
		if err != nil {
			t.Fatalf("unexpected error adding message: %v", err)
		}
	}

	msgs, err = store.GetRecentMessages(ctx, userID, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(msgs))
	}

	allMsgs, err := store.GetRecentMessages(ctx, userID, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(allMsgs) != 5 {
		t.Fatalf("expected 5 messages, got %d", len(allMsgs))
	}
}

func TestStore_Clear(t *testing.T) {
	ctx := context.Background()
	store := NewStore()

	u1 := int64(1)
	u2 := int64(2)

	_ = store.AddMessage(ctx, u1, "user", "hello")
	_ = store.AddMessage(ctx, u2, "user", "world")

	err := store.Clear(ctx, u1)
	if err != nil {
		t.Fatalf("unexpected error clearing: %v", err)
	}

	m1, _ := store.GetRecentMessages(ctx, u1, 10)
	if len(m1) != 0 {
		t.Fatalf("expected u1 history to be cleared")
	}

	m2, _ := store.GetRecentMessages(ctx, u2, 10)
	if len(m2) != 1 {
		t.Fatalf("expected u2 history to remain intact")
	}
}
