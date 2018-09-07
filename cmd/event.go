package cmd

import (
	"strconv"

	"github.com/mshindle/simdrone/bus"
	"github.com/mshindle/simdrone/events"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// handlerCmd represents the handler command
var eventCmd = &cobra.Command{
	Use:   "event",
	Short: "Event processor service for the drone army",
	Long: `
The event processor service is responsible for processing incoming events (as the name implies). In this scenario,
the processor will listen to the queues, do some processing on the event, and store it in the event store.

While an HTTP server is not necessary for this command, we create one as most health checkers work off a URL endpoint.
Plus it leaves a stub for us to work from later if we want to add more visibility.`,
	Run: runEventProcessor,
}

func init() {
	rootCmd.AddCommand(eventCmd)
	eventCmd.Flags().IntP("port", "p", 8081, "server port")
	eventCmd.Flags().String("bus", "amqp://guest:guest@firefly.dev:5672/", "bus / amqp connection string")
	viper.BindPFlag("port", eventCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("amqp.url", eventCmd.Flags().Lookup("bus"))
}

func runEventProcessor(cmd *cobra.Command, args []string) {
	config := &events.Config{
		DispatchConfig: bus.Config{
			URL: viper.GetString("amqp.url"),
		},
		AlertsQueueName: viper.GetString("queue.names.alert"),
		TelemetryQueueName: viper.GetString("queue.names.telemetry"),
		PositionsQueueName: viper.GetString("queue.names.position"),
	}
	server := events.NewServer(config)
	server.Run(":" + strconv.Itoa(viper.GetInt("port")))
}
