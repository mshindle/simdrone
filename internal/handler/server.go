package handler

import (
	"context"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/mshindle/simdrone/internal/bus"
	"github.com/mshindle/simdrone/internal/web"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
)

type Handler struct {
	e          *echo.Echo
	dispatcher bus.Dispatcher
	logger     zerolog.Logger
}

func New(dispatcher bus.Dispatcher, opts ...Option) *Handler {
	h := &Handler{
		dispatcher: dispatcher,
		e:          echo.New(),
		logger:     zerolog.Nop(),
	}
	for _, opt := range opts {
		opt(h)
	}

	h.e.Validator = web.NewStructValidator()
	h.e.HTTPErrorHandler = web.HTTPErrorHandler
	h.initRoutes()

	return h
}

func (h *Handler) initRoutes() {
	h.e.Use(
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
					h.logger.Info().
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
		web.AddContextLogger(h.logger, echo.HeaderXRequestID),
		middleware.Recover(),
	)

	apiGroup := h.e.Group("/api")
	cmdGroup := apiGroup.Group("/cmds")
	cmdGroup.POST("/alert", h.alertHandler)
	cmdGroup.POST("/telemetry", h.telemetryHandler)
	cmdGroup.POST("/position", h.positionHandler)
}

func (h *Handler) Run(ctx context.Context, addr string) error {
	sc := echo.StartConfig{
		Address:         addr,
		GracefulTimeout: 10 * time.Second,
		HideBanner:      true,
	}
	return sc.Start(ctx, h.e)
}

type Option func(h *Handler)

func WithLogger(logger zerolog.Logger) Option {
	return func(h *Handler) {
		h.logger = logger
	}
}

var Module = fx.Module("handler",
	fx.Provide(
		fx.Annotate(
			New,
			fx.ParamTags("", `group:"handlerOptions"`),
		),
	),
)

func AsOption(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(`group:"handlerOptions"`),
	)
}
