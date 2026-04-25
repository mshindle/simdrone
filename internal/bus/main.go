package bus

import "context"

// Exchange is the name of the topic for submitting events

type Dispatcher interface {
	Dispatch(ctx context.Context, topic string, message any) error
}

// Subscriber is transport-agnostic: NATS, RabbitMQ, local test bus, etc.
type Subscriber interface {
	Subscribe(ctx context.Context, topic string, handler MessageHandler) error
}

// MessageHandler handles raw transport payloads.
// Decoding into a domain event happens in the application layer.
type MessageHandler func(ctx context.Context, topic string, payload []byte) error
