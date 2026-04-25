package web

import (
	"context"

	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog"
)

// AddContextLogger creates middleware to inject a logger into the context
func AddContextLogger(logger zerolog.Logger, hdrRequestID string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			requestID := c.Response().Header().Get(hdrRequestID)
			ctx := logger.With().Str("request_id", requestID).Logger().WithContext(c.Request().Context())
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

func LoggerFromEchoContext(c *echo.Context) *zerolog.Logger {
	return LoggerFromContext(c.Request().Context())
}

func LoggerFromContext(c context.Context) *zerolog.Logger {
	return zerolog.Ctx(c)
}
