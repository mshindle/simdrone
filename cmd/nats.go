package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ipfans/fxlogger"
	"github.com/mshindle/simdrone/internal/bus/nats"
	"github.com/mshindle/simdrone/internal/config"
	"github.com/mshindle/simdrone/internal/event"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

var message string

// natsCmd represents the nats command
var natsCmd = &cobra.Command{
	Use:   "nats",
	Short: "Send a Pinged event to a NATS server",
	Long:  ``,
}

var natsPingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Send a Pinged event to a NATS server",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fx.New(
			fx.WithLogger(
				func(logger zerolog.Logger) fxevent.Logger {
					return fxlogger.WithZerolog(logger)()
				}),
			commonModule(cmd),
			config.Module,
			nats.Module,
			fx.Invoke(natsSend),
		).Run()
	},
}

var natsInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Send a Pinged event to a NATS server",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fx.New(
			fx.WithLogger(
				func(logger zerolog.Logger) fxevent.Logger {
					return fxlogger.WithZerolog(logger)()
				}),
			commonModule(cmd),
			config.Module,
			nats.Module,
			fx.Invoke(natsInit),
		).Run()
	},
}

func init() {
	rootCmd.AddCommand(natsCmd)

	natsCmd.AddCommand(natsPingCmd)
	natsPingCmd.Flags().StringVarP(&message, "message", "m", "ping!", "message to send")

	natsCmd.AddCommand(natsInitCmd)
}

func natsSend(lc fx.Lifecycle, m *nats.Messenger, ctx context.Context, l zerolog.Logger, shutdowner fx.Shutdowner) {
	ctxCancel, cancel := context.WithCancel(ctx)
	lc.Append(fx.Hook{
		OnStart: func(startCtx context.Context) error {
			// run a subscriber
			var wg sync.WaitGroup
			err := m.Subscribe(ctxCancel, event.Ping, func(ctx context.Context, _ string, payload []byte) error {
				fmt.Printf("Received message on [%s]: %s\n", event.Ping, string(payload))
				wg.Done()
				return nil
			})

			// Build the message
			var packet []byte
			packet, err = json.Marshal(event.Pinged{Message: message, ReceivedAt: time.Now()})
			if err != nil {
				l.Error().Err(err).Msg("failed to marshal pinged event")
				return err
			}
			fmt.Printf("Sending message on [%s]: %s\n", event.Ping, string(packet))

			// send the message
			wg.Add(1)
			err = m.Publish(event.Ping, packet)
			if err != nil {
				l.Error().Err(err).Msg("failed to publish pinged event")
				return err
			}
			l.Info().Str("topic", event.Ping).Msg("message sent")
			err = m.Flush()
			if err != nil {
				l.Error().Err(err).Msg("failed to flush NATS connection")
				return err
			}

			// wait for the subscriber to receive the message
			wg.Wait()
			l.Info().Msg("message sent and received")

			return shutdowner.Shutdown()
		},
		OnStop: func(stopCtx context.Context) error {
			l.Info().Msg("shutting down NATS client")
			cancel()
			_ = m.Unsubscribe(event.Ping)
			_ = m.Close()
			return nil
		},
	})
}

func natsInit(lc fx.Lifecycle, m *nats.Messenger, l zerolog.Logger, shutdowner fx.Shutdowner) {
	lc.Append(fx.Hook{
		OnStart: func(startCtx context.Context) error {
			err := m.Initialize()
			if err != nil {
				l.Error().Err(err).Msg("failed to initialize NATS stream")
				return err
			}
			l.Info().Msg("NATS stream initialized")
			return shutdowner.Shutdown()
		},
		OnStop: func(stopCtx context.Context) error {
			return m.Close()
		},
	})
}
