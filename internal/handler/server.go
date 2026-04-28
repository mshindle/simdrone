package handler

import (
	"context"
	"time"

	"github.com/labstack/echo-opentelemetry"
	"github.com/labstack/echo/v5"
	"github.com/mshindle/simdrone/internal/bus"
	"github.com/mshindle/simdrone/internal/web"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/fx"
)

type Handler struct {
	e          *echo.Echo
	dispatcher bus.Dispatcher
	logger     zerolog.Logger
	tp         trace.TracerProvider
	tracer     trace.Tracer
}

func New(dispatcher bus.Dispatcher, opts ...Option) *Handler {
	h := &Handler{
		dispatcher: dispatcher,
		e:          echo.New(),
		logger:     zerolog.Nop(),
		tp:         noop.NewTracerProvider(),
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
		web.CommonMiddleware(
			h.logger,
			echootel.Config{
				TracerProvider: h.tp,
			},
		)...,
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

func WithTraceProvider(tp trace.TracerProvider) Option {
	return func(h *Handler) {
		h.tp = tp
	}
}

func WithTracer(t trace.Tracer) Option {
	return func(h *Handler) {
		h.tracer = t
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
