package cmd

import (
	"context"
	"errors"

	"github.com/ipfans/fxlogger"
	"github.com/mshindle/simdrone/internal/bus"
	"github.com/mshindle/simdrone/internal/bus/nats"
	"github.com/mshindle/simdrone/internal/config"
	"github.com/mshindle/simdrone/internal/event"
	"github.com/mshindle/simdrone/internal/evtproc"
	"github.com/mshindle/simdrone/internal/repository/mongodb"
	"github.com/mshindle/simdrone/internal/telemetry"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

const procTracerName = "processor"

// eventCmd represents the event processor command
var eventCmd = &cobra.Command{
	Use:   "process",
	Short: "Event processor service for the drone army",
	Long: `
The event processor service is responsible for processing incoming events (as the name implies). In this scenario,
the processor will listen to the queues, do some processing on the event, and store it in the event store.

While an HTTP server is not necessary for this command, we create one as most health checkers work off a URL endpoint.
Plus it leaves a stub for us to work from later if we want to add more visibility.`,
	Run: func(cmd *cobra.Command, args []string) {
		fx.New(
			fx.WithLogger(
				func(logger zerolog.Logger) fxevent.Logger {
					return fxlogger.WithZerolog(logger)()
				}),
			commonModule(cmd),
			config.Module,
			nats.Module,
			telemetry.Module,
			fx.Provide(
				func() telemetry.TracerName { return procTracerName },
			),
			evtproc.Module,
			mongodb.Module,
			fx.Provide(
				func(server *evtproc.Server) WebServer { return server },
				evtproc.AsOption(evtproc.WithLogger),
			),
			fx.Provide(
				func(m *nats.Messenger) bus.Subscriber { return m },
			),
			fx.Invoke(
				func(lc fx.Lifecycle, ctx context.Context, w WebServer, l zerolog.Logger, cfg *config.Config) {
					invokeWebServer(lc, ctx, w, l, cfg.Processor.Port)
				},
				runProcessors,
			),
		).Run()
	},
}

func init() {
	rootCmd.AddCommand(eventCmd)
}

func runProcessors(lc fx.Lifecycle, ctx context.Context, tracer trace.Tracer, subscriber bus.Subscriber, r *mongodb.EventRollupRepository, l zerolog.Logger) {
	ctxCancel, cancel := context.WithCancel(ctx)

	lc.Append(fx.Hook{
		OnStart: func(startCtx context.Context) error {
			setups := []func() error{
				func() error {
					return evtproc.DispatchEvent(ctxCancel, tracer, subscriber, event.AlertSignal, r.AddAlert)
				},
				func() error {
					return evtproc.DispatchEvent(ctxCancel, tracer, subscriber, event.TelemetryUpdate, r.AddTelemetry)
				},
				func() error {
					return evtproc.DispatchEvent(ctxCancel, tracer, subscriber, event.PositionUpdate, r.AddPosition)
				},
			}
			var errs error
			for _, setup := range setups {
				if err := setup(); err != nil {
					errs = errors.Join(errs, err)
				}

			}
			l.Debug().Err(errs).Msg("finish subscribe process")
			return errs
		},
		OnStop: func(stopCtx context.Context) error {
			cancel()
			return nil
		},
	})
}
