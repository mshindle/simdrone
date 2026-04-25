package evtproc

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/mshindle/simdrone/internal/web"
	"github.com/rs/zerolog"
)

type Server struct {
	e      *echo.Echo
	logger zerolog.Logger
}

func NewServer(opts ...Option) *Server {
	server := &Server{
		e:      echo.New(),
		logger: zerolog.Nop(),
	}
	for _, opt := range opts {
		opt(server)
	}

	server.e.Validator = web.NewStructValidator()
	server.e.HTTPErrorHandler = web.HTTPErrorHandler
	server.initRoutes()

	return server
}

func (ep *Server) initRoutes() {
	ep.e.Use(
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
					ep.logger.Info().
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
		web.AddContextLogger(ep.logger, echo.HeaderXRequestID),
		middleware.Recover(),
	)

	ep.e.GET("/", ep.homeHandler)
}

func (ep *Server) Run(ctx context.Context, addr string) error {
	sc := echo.StartConfig{
		Address:         addr,
		GracefulTimeout: 10 * time.Second,
		HideBanner:      true,
	}
	return sc.Start(ctx, ep.e)
}

func (ep *Server) homeHandler(c *echo.Context) error {
	return c.HTML(http.StatusOK, "<h1>Simdrone- Event Processor see http://github.com/mshindle/simdrone</h1>")
}

type Option func(s *Server)

func WithLogger(logger zerolog.Logger) Option {
	return func(s *Server) {
		s.logger = logger
	}
}
