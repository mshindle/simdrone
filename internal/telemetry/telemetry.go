package telemetry

import (
	"context"
	"errors"

	"github.com/mshindle/simdrone/internal/config"
	"github.com/mshindle/structures/ringbuffer"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
)

type ServiceName string
type TracerName string

const (
	AttrSubject  = "bus.subject"
	AttrDBSystem = "db.system"
)

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

func NewTracerProvider(ctx context.Context, exporter sdktrace.SpanExporter, serviceName ServiceName) (*sdktrace.TracerProvider, error) {
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

func NewTracer(tp trace.TracerProvider, tracerName TracerName) trace.Tracer {
	return tp.Tracer(string(tracerName))
}

// Module provides telemetry functionality for the application. The caller of this module must provide the TracerName.
var Module = fx.Module("telemetry",
	fx.Provide(
		func(ctx context.Context, cfg *config.Config) (sdktrace.SpanExporter, error) {
			var exporter sdktrace.SpanExporter
			var err error

			switch cfg.Telemetry.Exporter {
			case "xray":
				// Placeholder for our next step!
				return nil, errors.New("xray exporter not yet implemented")
			case "jaeger":
				fallthrough
			default:
				// Create the standard OTLP gRPC exporter
				exporter, err = otlptracegrpc.New(ctx,
					otlptracegrpc.WithInsecure(),
					otlptracegrpc.WithEndpoint(cfg.Telemetry.Endpoint),
				)
			}
			if err != nil {
				return nil, err
			}

			return exporter, nil
		},
		func(cfg *config.Config) ServiceName { return ServiceName(cfg.Telemetry.Name) },
		fx.Annotate(
			NewTracerProvider,
			fx.As(new(trace.TracerProvider)),
			fx.As(fx.Self()),
		),
		NewTracer,
	),
	fx.Invoke(
		func(lc fx.Lifecycle, tp *sdktrace.TracerProvider) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					otel.SetTracerProvider(tp)
					// Add this line to enable Header Injection/Extraction!
					otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
					return nil
				},
				OnStop: func(ctx context.Context) error {
					return tp.Shutdown(ctx) // Ensure spans are flushed on exit
				},
			})
		},
	),
)
