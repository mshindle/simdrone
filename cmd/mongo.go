package cmd

import (
	"context"
	"fmt"

	"github.com/ipfans/fxlogger"
	"github.com/mshindle/simdrone/internal/bus/nats"
	"github.com/mshindle/simdrone/internal/config"
	"github.com/mshindle/simdrone/internal/repository/mongodb"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

// mongoCmd represents the mongo command
var mongoCmd = &cobra.Command{
	Use:   "mongo",
	Short: "Commands to interact with MongoDB",
	Long:  ``,
}

var mongoInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize MongoDB with required indices",
	Long:  `Initialize the MongoDB database by creating the required indices for the event rollup repository.`,
	Run: func(cmd *cobra.Command, args []string) {
		fx.New(
			fx.WithLogger(
				func(logger zerolog.Logger) fxevent.Logger {
					return fxlogger.WithZerolog(logger)()
				}),
			commonModule(cmd),
			config.Module,
			nats.Module,
			mongodb.Module,
			fx.Invoke(
				mongoInit,
			),
		).Run()
	},
}

func init() {
	rootCmd.AddCommand(mongoCmd)
	mongoCmd.AddCommand(mongoInitCmd)
}

func mongoInit(lc fx.Lifecycle, l zerolog.Logger, e *mongodb.EventRollupRepository, shutdowner fx.Shutdowner) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			l.Info().Msg("creating indices...")
			err := mongodb.EnsureIndexes(ctx, e)
			if err != nil {
				l.Error().Err(err).Msg("failed to create indices")
				return fmt.Errorf("failed to create indices: %w", err)
			}

			return shutdowner.Shutdown()
		},
	})
}
