package bus

import (
	"context"
	"sync"
)

// LocalDispatcher provides a simple in-memory dispatcher for testing
// and local development.
type LocalDispatcher struct {
	mu       sync.RWMutex
	messages map[string][]any
}

// NewLocalDispatcher instantiates a new LocalDispatcher
func NewLocalDispatcher() *LocalDispatcher {
	return &LocalDispatcher{
		messages: make(map[string][]any),
	}
}

// Dispatch fakes sending a message
func (l *LocalDispatcher) Dispatch(_ context.Context, routingKey string, message any) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.messages[routingKey] = append(l.messages[routingKey], message)
	return nil
}

// GetMessages returns all messages dispatched for a specific key
func (l *LocalDispatcher) GetMessages(routingKey string) []any {
	l.mu.RLock()
	defer l.mu.RUnlock()

	m, ok := l.messages[routingKey]
	if !ok {
		return nil
	}
	// Return a copy to avoid concurrent access issues with the internal slice
	res := make([]any, len(m))
	copy(res, m)
	return res
}

// Clear removes all messages for all keys
func (l *LocalDispatcher) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.messages = make(map[string][]any)
}
