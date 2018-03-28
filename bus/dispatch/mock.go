package dispatch

// MockDispatcher provides a string mapping so we can assert on which Messages
// have been dispatched by calling clients.
type mockDispatcher struct {
	Messages []interface{}
}

// NewMockDispatcher creates a fake dispatcher
func NewMockDispatcher() *mockDispatcher {
	m := &mockDispatcher{}
	m.Messages = make([]interface{}, 0)
	return m
}

// Dispatch fakes sending a message
func (m *mockDispatcher) Dispatch(routingKey string, message interface{}) error {
	m.Messages = append(m.Messages, message)
	return nil
}
