package telemetry

import (
	"context"
	"time"
)

// detachedContext exposes the values of the parent context (like OTel SpanContext)
// but hides its cancellation signals.
type detachedContext struct {
	context.Context
}

func (detachedContext) Deadline() (deadline time.Time, ok bool) { return }
func (detachedContext) Done() <-chan struct{}                   { return nil }
func (detachedContext) Err() error                              { return nil }

// Detach Context preserves OpenTelemetry and Logger values from an HTTP request
// but prevents background workers from being canceled when the HTTP response is sent.
func DetachContext(ctx context.Context) context.Context {
	return detachedContext{ctx}
}
