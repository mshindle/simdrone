package cmd

import (
	"context"

	"github.com/ipfans/fxlogger"
	"github.com/mshindle/simdrone/internal/bus/nats"
	"github.com/mshindle/simdrone/internal/config"
	"github.com/mshindle/simdrone/internal/repository/mongodb"
	"github.com/mshindle/simdrone/internal/view"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

// viewCmd represents the view command
var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "Query and display the latest drone data",
	Long: `
The view command provides a real-time perspective of the drone army's current state. 
It queries the system to retrieve and display the most recent telemetry, position, 
and alert data from all active drones.

This service acts as the 'eyes' of the operation, allowing operators to monitor 
drone health, locations, and any emergency signals dispatched to the network.`,
	Run: func(cmd *cobra.Command, args []string) {
		fx.New(
			fx.WithLogger(
				func(logger zerolog.Logger) fxevent.Logger {
					return fxlogger.WithZerolog(logger)()
				}),
			commonModule(cmd),
			config.Module,
			nats.Module,
			view.Module,
			mongodb.Module,
			fx.Provide(
				func(r *mongodb.EventRollupRepository) view.EventRepository { return r },
				func(srv *view.Server) WebServer { return srv },
				view.AsOption(view.WithLogger),
				view.AsOption(view.WithRepository),
			),
			fx.Invoke(
				func(lc fx.Lifecycle, ctx context.Context, w WebServer, l zerolog.Logger, cfg *config.Config) {
					invokeWebServer(lc, ctx, w, l, cfg.View.Port)
				},
			),
		).Run()
	},
}

func init() {
	rootCmd.AddCommand(viewCmd)
}
