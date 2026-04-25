package evtproc

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/mshindle/simdrone/internal/bus"
	"github.com/mshindle/simdrone/internal/event"
)

type mockSubscriber struct {
	handler bus.MessageHandler
	topic   string
}

func (m *mockSubscriber) Subscribe(_ context.Context, topic string, handler bus.MessageHandler) error {
	m.topic = topic
	m.handler = handler
	return nil
}

func TestDispatchEvent(t *testing.T) {
	ctx := context.Background()
	topic := "test.topic"

	t.Run("successful dispatch", func(t *testing.T) {
		sub := &mockSubscriber{}
		processed := false
		var receivedEvt *event.TelemetryUpdated

		processor := func(ctx context.Context, evt *event.TelemetryUpdated) error {
			processed = true
			receivedEvt = evt
			return nil
		}

		err := DispatchEvent(ctx, sub, topic, processor)
		if err != nil {
			t.Fatalf("DispatchEvent failed: %v", err)
		}

		if sub.topic != topic {
			t.Errorf("expected topic %s, got %s", topic, sub.topic)
		}

		if sub.handler == nil {
			t.Fatal("expected handler to be registered")
		}

		// Simulate message receipt
		evt := event.TelemetryUpdated{DroneID: "drone1", RemainingBattery: 50}
		payload, _ := json.Marshal(evt)
		err = sub.handler(ctx, topic, payload)
		if err != nil {
			t.Errorf("handler failed: %v", err)
		}

		if !processed {
			t.Error("processor was not called")
		}
		if receivedEvt.DroneID != "drone1" {
			t.Errorf("expected drone1, got %s", receivedEvt.DroneID)
		}
	})

	t.Run("unmarshal error", func(t *testing.T) {
		sub := &mockSubscriber{}
		processor := func(ctx context.Context, evt *event.TelemetryUpdated) error {
			return nil
		}

		_ = DispatchEvent(ctx, sub, topic, processor)

		err := sub.handler(ctx, topic, []byte("invalid json"))
		if err == nil {
			t.Error("expected error due to invalid json, got nil")
		}
	})

	t.Run("processor error", func(t *testing.T) {
		sub := &mockSubscriber{}
		expectedErr := errors.New("processor failed")
		processor := func(ctx context.Context, evt *event.TelemetryUpdated) error {
			return expectedErr
		}

		_ = DispatchEvent(ctx, sub, topic, processor)

		evt := event.TelemetryUpdated{DroneID: "drone1"}
		payload, _ := json.Marshal(evt)
		err := sub.handler(ctx, topic, payload)
		if !errors.Is(err, expectedErr) {
			t.Errorf("expected error %v, got %v", expectedErr, err)
		}
	})
}
