package web

import (
	echootel "github.com/labstack/echo-opentelemetry"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

func CommonMiddleware(l zerolog.Logger, config echootel.Config) []echo.MiddlewareFunc {
	return []echo.MiddlewareFunc{
		echootel.NewMiddlewareWithConfig(config),
		middleware.RequestID(),
		middleware.RequestLoggerWithConfig(
			middleware.RequestLoggerConfig{
				LogRequestID: true,
				LogHost:      true,
				LogURI:       true,
				LogStatus:    true,
				LogMethod:    true,
				LogLatency:   true,
				LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
					// 1. Extract the active span context from the request
					spanCtx := trace.SpanFromContext(c.Request().Context()).SpanContext()

					// 2. Build the base log event
					logEvent := l.Info().
						Int("status", v.Status).
						Str("host", v.Host).
						Str("method", v.Method).
						Str("uri", v.URI).
						Str("request_id", v.RequestID).
						Dur("latency", v.Latency)

					if spanCtx.IsValid() {
						logEvent.
							Str("trace_id", spanCtx.TraceID().String()).
							Str("span_id", spanCtx.SpanID().String())
					}

					logEvent.Msg("request")
					return nil
				},
			},
		),
		AddContextLogger(l, echo.HeaderXRequestID),
		middleware.Recover(),
	}
}
