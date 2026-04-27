package telemetry

import (
	"context"

	"github.com/mshindle/structures/ringbuffer"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
)

type ServiceName string

type TracedRingBuffer[T any] struct {
	*ringbuffer.RingBuffer[T]
	tracer trace.Tracer
}

// NewTracedRingBuffer is a factory for Fx injection
func NewTracedRingBuffer[T any](capacity int, tracer trace.Tracer) *TracedRingBuffer[T] {
	return &TracedRingBuffer[T]{
		RingBuffer: ringbuffer.New[T](capacity),
		tracer:     tracer,
	}
}

// OverwritePush instrumented for telemetry
func (trb *TracedRingBuffer[T]) OverwritePush(ctx context.Context, v T) {
	var span trace.Span
	ctx, span = trb.tracer.Start(ctx, "RingBuffer.OverwritePush")
	defer span.End()

	// Capture the length before push to see how full the buffer is
	span.SetAttributes(attribute.Int("buffer.length_before", trb.Len()))

	trb.RingBuffer.OverwritePush(v)
}

func (trb *TracedRingBuffer[T]) Push(ctx context.Context, v T) error {
	ctx, span := trb.tracer.Start(ctx, "RingBuffer.Push")
	defer span.End()

	err := trb.RingBuffer.Push(v)
	if err != nil {
		span.RecordError(err) // Record failures if buffer is full
	}
	return err
}

func NewTracerProvider(ctx context.Context, serviceName ServiceName) (*sdktrace.TracerProvider, error) {
	// Create an OTLP exporter (targets a collector or Jaeger)
	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	// Identify this service
	res, err := resource.New(ctx,
		resource.WithAttributes(semconv.ServiceNameKey.String(string(serviceName))),
	)
	if err != nil {
		return nil, err
	}

	// Set up the provider with a batch processor for performance
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	return tp, nil
}

var Module = fx.Module("telemetry",
	fx.Provide(
		NewTracerProvider,
	),
	fx.Invoke(
		func(lc fx.Lifecycle, tp *sdktrace.TracerProvider) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					otel.SetTracerProvider(tp)
					return nil
				},
				OnStop: func(ctx context.Context) error {
					return tp.Shutdown(ctx) // Ensure spans are flushed on exit
				},
			})
		},
	),
)
