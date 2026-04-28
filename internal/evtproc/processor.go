package evtproc

import (
	"context"
	"encoding/json"

	"github.com/mshindle/simdrone/internal/bus"
	"github.com/mshindle/simdrone/internal/event"
	"github.com/mshindle/simdrone/internal/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type DroneProcessor[T event.DroneEvents] func(ctx context.Context, evt *T) error

func DispatchEvent[T event.DroneEvents](ctx context.Context, tracer trace.Tracer, subscriber bus.Subscriber, topic string, p DroneProcessor[T]) error {
	return subscriber.Subscribe(ctx, topic, func(cx context.Context, _ string, payload []byte) error {
		var evt T

		cx, span := tracer.Start(cx, "EventProcessor.Process")
		defer span.End()
		span.SetAttributes(attribute.String(telemetry.AttrSubject, topic))

		err := json.Unmarshal(payload, &evt)
		if err != nil {
			span.RecordError(err)
			return err
		}
		return p(cx, &evt)
	})
}
