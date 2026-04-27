package nats

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/mshindle/simdrone/internal/bus"
	"github.com/mshindle/simdrone/internal/config"
	"github.com/mshindle/simdrone/internal/telemetry"
	"github.com/mshindle/simdrone/internal/web"
	"github.com/nats-io/nats.go"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/fx"
)

const (
	msgIDHeader = "msg_id"
)

type Messenger struct {
	conn        *nats.Conn
	js          nats.JetStreamContext
	l           zerolog.Logger
	subscribers map[string]*nats.Subscription
	mu          sync.RWMutex
}

func NewMessenger(conn *nats.Conn, js nats.JetStreamContext, logger zerolog.Logger) *Messenger {
	return &Messenger{
		conn:        conn,
		js:          js,
		l:           logger,
		subscribers: make(map[string]*nats.Subscription),
	}
}

// Publish publishes a message to a NATS subject using JetStream.
func (m *Messenger) Publish(subject string, data []byte) error {
	return m.conn.Publish(subject, data)
}

func (m *Messenger) Close() error {
	m.conn.Close()
	return nil
}

// Dispatch wraps NATS publish method to work with the commonly defined bus interfaces.
func (m *Messenger) Dispatch(ctx context.Context, routingKey string, message any) error {
	var err error
	var data []byte
	asyncCtx := telemetry.DetachContext(ctx)

	l := web.LoggerFromContext(asyncCtx)
	hdr := nats.Header{}
	hdr.Set(msgIDHeader, xid.New().String())
	hdr.Set("source", "cmdHandler")

	// 2. INJECT the OpenTelemetry Trace Context into the NATS headers.
	// We cast nats.Header to http.Header since they share the same underlying map[string][]string type.
	otel.GetTextMapPropagator().Inject(asyncCtx, propagation.HeaderCarrier(hdr))

	switch message.(type) {
	case []byte:
		data = message.([]byte)
	case string:
		data = []byte(message.(string))
	default:
		data, err = json.Marshal(message)
		if err != nil {
			return err
		}
	}
	msg := new(nats.Msg{Header: hdr, Subject: routingKey, Data: data})
	l.Debug().Any("msg", msg).Msg("publishing message to NATS")
	return m.conn.PublishMsg(msg)
}

func (m *Messenger) Subscribe(ctx context.Context, topic string, handler bus.MessageHandler) error {
	m.mu.Lock()
	if _, ok := m.subscribers[topic]; ok {
		m.mu.Unlock()
		return fmt.Errorf("subscriber for topic %s already exists", topic)
	}
	m.mu.Unlock()

	ncs, err := m.conn.Subscribe(topic, func(msg *nats.Msg) {
		// EXTRACT the OpenTelemetry Context from the incoming NATS headers.
		// We use the background context `ctx` passed to Subscribe as the root.
		msgCtx := otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(msg.Header))
		l := m.l.With().Str("msg_id", msg.Header.Get(msgIDHeader)).Str("subject", msg.Subject).Logger()
		l.Debug().Msg("received message")

		errh := handler(msgCtx, msg.Subject, msg.Data)
		if errh != nil {
			l.Error().Err(errh).Msg("error handling event")
			// Negative Acknowledgment: Tell JetStream to redeliver the message later
			_ = msg.Nak()
			return
		}
		l.Debug().Msg("acknowledged message")
		_ = msg.AckSync()
	})
	if err != nil {
		return err
	}

	m.mu.Lock()
	m.subscribers[topic] = ncs
	m.mu.Unlock()

	// Spin up a goroutine to clean up the subscription when the context is done
	go func() {
		<-ctx.Done()
		_ = m.Unsubscribe(topic)
	}()

	return nil
}

func (m *Messenger) Unsubscribe(topic string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sub, ok := m.subscribers[topic]
	if !ok {
		return fmt.Errorf("no subscriber for topic %s", topic)
	}

	err := sub.Unsubscribe()
	delete(m.subscribers, topic)
	return err
}

func (m *Messenger) Flush() error {
	return m.conn.Flush()
}

func (m *Messenger) Initialize() error {
	var errs error
	var stream = map[string]string{
		"DRONE_EVENTS":  "events.drone.>",
		"SYSTEM_EVENTS": "events.system.>",
	}

	for name, pattern := range stream {
		_, err := m.js.AddStream(&nats.StreamConfig{
			Name:     name,
			Subjects: []string{pattern},
		})
		errs = errors.Join(errs, err)
	}

	return errs
}

var Module = fx.Module("nats",
	fx.Provide(func(cfg *config.Config, logger zerolog.Logger) (*Messenger, error) {
		l := logger.With().Str("server", cfg.Nats.URL).Logger()
		// Connect to NATS server
		nc, err := nats.Connect(cfg.Nats.URL)
		if err != nil {
			l.Error().Err(err).Msg("failed to connect to NATS server")
			return nil, err
		}
		l.Info().Msg("connected to NATS server")

		var js nats.JetStreamContext
		js, err = nc.JetStream()
		if err != nil {
			l.Error().Err(err).Msg("failed to connect to NATS JetStream")
			return nil, err
		}
		l.Info().Msg("connected to NATS JetStream")
		return NewMessenger(nc, js, logger), nil
	}),
)
