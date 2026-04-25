// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"

	"github.com/ipfans/fxlogger"
	"github.com/mshindle/simdrone/internal/bus"
	"github.com/mshindle/simdrone/internal/bus/nats"
	"github.com/mshindle/simdrone/internal/config"
	"github.com/mshindle/simdrone/internal/handler"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

var useLocal bool

// handlerCmd represents the handler command
var handlerCmd = &cobra.Command{
	Use:   "handler",
	Short: "Service to handle and dispatch drone commands",
	Long: `
The command handler service is responsible for processing incoming commands, converting them into 
an endpoint-agnostic event, and submitting them to the appropriate queue. Once submitted, the command
handler washes its _hands_ of the request and moves on to the next task.

This service acts as a gateway for the drone army, ensuring that commands are 
properly formatted and queued before being handled by specialized workers.`,
	Run: func(cmd *cobra.Command, args []string) {
		fx.New(
			fx.WithLogger(
				func(logger zerolog.Logger) fxevent.Logger {
					return fxlogger.WithZerolog(logger)()
				}),
			commonModule(cmd),
			config.Module,
			nats.Module,
			configureDispatcher(),
			handler.Module,
			fx.Provide(
				func(h *handler.Handler) WebServer { return h },
				handler.AsOption(handler.WithLogger),
			),
			fx.Invoke(
				func(lc fx.Lifecycle, ctx context.Context, w WebServer, l zerolog.Logger, cfg *config.Config) {
					invokeWebServer(lc, ctx, w, l, cfg.Handler.Port)
				},
			),
		).Run()
	},
}

func init() {
	rootCmd.AddCommand(handlerCmd)
	handlerCmd.Flags().IntP("port", "p", 8080, "server port")
	handlerCmd.Flags().BoolVar(&useLocal, "local", false, "use local dispatcher")

	_ = v.BindPFlag("handler.port", handlerCmd.Flags().Lookup("port"))
}

func configureDispatcher() fx.Option {
	var f any

	f = func(m *nats.Messenger) bus.Dispatcher { return m }
	if useLocal {
		f = func() bus.Dispatcher { return bus.NewLocalDispatcher() }
	}
	return fx.Provide(f)
}
