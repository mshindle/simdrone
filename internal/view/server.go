package view

import (
	"context"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/mshindle/simdrone/internal/event"
	"github.com/mshindle/simdrone/internal/web"
	"github.com/rs/zerolog"
	"go.uber.org/fx"
)

type Server struct {
	e      *echo.Echo
	logger zerolog.Logger
	repo   EventRepository
}

func New(opts ...Option) *Server {
	srv := &Server{
		e:      echo.New(),
		logger: zerolog.Nop(),
		repo:   &mockEventRepository{},
	}
	for _, opt := range opts {
		opt(srv)
	}

	srv.e.Validator = web.NewStructValidator()
	srv.e.HTTPErrorHandler = web.HTTPErrorHandler
	srv.initRoutes()

	return srv
}

func (srv *Server) initRoutes() {
	srv.e.Use(web.CommonMiddleware(srv.logger)...)

	dronesGroup := srv.e.Group("/drones")
	dronesGroup.GET("", srv.activeHandler)
	dronesGroup.GET("/:droneID/lastAlert", lastEventHandler[event.AlertSignalled](srv.repo.GetAlert))
	dronesGroup.GET("/:droneID/lastTelemetry", lastEventHandler[event.TelemetryUpdated](srv.repo.GetTelemetry))
	dronesGroup.GET("/:droneID/lastPosition", lastEventHandler[event.PositionChanged](srv.repo.GetPosition))
}

func (srv *Server) Run(ctx context.Context, addr string) error {
	sc := echo.StartConfig{
		Address:         addr,
		GracefulTimeout: 10 * time.Second,
		HideBanner:      true,
	}
	return sc.Start(ctx, srv.e)
}

type Option func(srv *Server)

func WithLogger(logger zerolog.Logger) Option {
	return func(srv *Server) {
		srv.logger = logger
	}
}

func WithRepository(repo EventRepository) Option {
	return func(srv *Server) {
		srv.repo = repo
	}
}

var Module = fx.Module("viewer",
	fx.Provide(
		fx.Annotate(
			New,
			fx.ParamTags(`group:"viewOptions"`),
		),
	),
)

func AsOption(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(`group:"viewOptions"`),
	)
}
