package bus

import (
	"context"
	"testing"
)

func TestNewLocalDispatcher(t *testing.T) {
	d := NewLocalDispatcher()
	if d == nil {
		t.Fatal("Expected NewLocalDispatcher to return a non-nil value")
	}
	if d.messages == nil {
		t.Error("Expected d.messages to be initialized")
	}
}

func TestLocalDispatcher_Dispatch(t *testing.T) {
	d := NewLocalDispatcher()
	ctx := context.Background()

	t.Run("single message", func(t *testing.T) {
		key := "test.key"
		msg := "hello world"
		err := d.Dispatch(ctx, key, msg)
		if err != nil {
			t.Errorf("Dispatch failed: %v", err)
		}

		msgs := d.GetMessages(key)
		if len(msgs) != 1 {
			t.Errorf("Expected 1 message, got %d", len(msgs))
		}
		if msgs[0] != msg {
			t.Errorf("Expected message %v, got %v", msg, msgs[0])
		}
	})

	t.Run("multiple messages for same key", func(t *testing.T) {
		key := "test.key2"
		msg1 := "msg 1"
		msg2 := "msg 2"
		_ = d.Dispatch(ctx, key, msg1)
		_ = d.Dispatch(ctx, key, msg2)

		msgs := d.GetMessages(key)
		if len(msgs) != 2 {
			t.Errorf("Expected 2 messages, got %d", len(msgs))
		}
		if msgs[0] != msg1 || msgs[1] != msg2 {
			t.Errorf("Messages out of order or incorrect: %v", msgs)
		}
	})

	t.Run("messages for different keys", func(t *testing.T) {
		key1 := "key1"
		key2 := "key2"
		_ = d.Dispatch(ctx, key1, "val1")
		_ = d.Dispatch(ctx, key2, "val2")

		if len(d.GetMessages(key1)) != 1 {
			t.Errorf("Expected 1 message for key1")
		}
		if len(d.GetMessages(key2)) != 1 {
			t.Errorf("Expected 1 message for key2")
		}
	})

	t.Run("Clear", func(t *testing.T) {
		d.Clear()
		if len(d.GetMessages("key1")) != 0 {
			t.Errorf("Expected 0 messages after Clear")
		}
	})
}
