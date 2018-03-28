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
	"strconv"

	"github.com/mshindle/simdrone/bus"
	"github.com/mshindle/simdrone/handler"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// handlerCmd represents the handler command
var handlerCmd = &cobra.Command{
	Use:   "handler",
	Short: "Command handler service for the drone army",
	Long: `
The command handler service is responsible for processing incoming commands and
converting them into events. The events will be dispatched into the messaging system.`,
	Run: runHandler,
}

func init() {
	rootCmd.AddCommand(handlerCmd)
	handlerCmd.Flags().IntP("port", "p", 8080, "server port")
	handlerCmd.Flags().String("amqp_conn", "amqp://guest:guest@firefly.dev:5672/", "amqp connection string")
	viper.BindPFlag("port", handlerCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("amqp.url", handlerCmd.Flags().Lookup("amqp_conn"))
}

func runHandler(cmd *cobra.Command, args []string) {
	config := &handler.Config{
		DispatchConfig: bus.Config{
			URL: viper.GetString("amqp.url"),
		},
	}
	server := handler.NewServer(config)
	server.Run(":" + strconv.Itoa(viper.GetInt("port")))
}
