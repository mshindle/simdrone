package evtproc

import (
	"context"
	"encoding/json"

	"github.com/mshindle/simdrone/internal/bus"
	"github.com/mshindle/simdrone/internal/event"
)

type DroneProcessor[T event.DroneEvents] func(ctx context.Context, evt *T) error

func DispatchEvent[T event.DroneEvents](ctx context.Context, subscriber bus.Subscriber, topic string, p DroneProcessor[T]) error {
	return subscriber.Subscribe(ctx, topic, func(cx context.Context, _ string, payload []byte) error {
		var evt T
		err := json.Unmarshal(payload, &evt)
		if err != nil {
			return err
		}
		return p(cx, &evt)
	})
}
