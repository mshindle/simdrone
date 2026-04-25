package web

import (
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/rs/zerolog"
)

func CommonMiddleware(l zerolog.Logger) []echo.MiddlewareFunc {
	return []echo.MiddlewareFunc{
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
					l.Info().
						Int("status", v.Status).
						Str("host", v.Host).
						Str("method", v.Method).
						Str("uri", v.URI).
						Str("request_id", v.RequestID).
						Dur("latency", v.Latency).
						Msg("request")
					return nil
				},
			},
		),
		AddContextLogger(l, echo.HeaderXRequestID),
		middleware.Recover(),
	}
}
