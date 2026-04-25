package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/rs/zerolog"
	"go.uber.org/fx"
)

func invokeWebServer(lc fx.Lifecycle, ctx context.Context, w WebServer, l zerolog.Logger, port int) {
	serverCtx, cancel := context.WithCancel(ctx)
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			address := fmt.Sprintf(":%d", port)
			l.Info().Str("address", address).Msg("server starting")
			go func(a string) {
				if err := w.Run(serverCtx, a); err != nil && !errors.Is(err, http.ErrServerClosed) {
					l.Error().Err(err).Msg("api service failed")
				}
			}(address)
			return nil
		},
		OnStop: func(_ context.Context) error {
			l.Info().Msg("shutting down HTTP server")
			cancel()
			return nil
		},
	})
}

type WebServer interface {
	Run(ctx context.Context, address string) error
}
