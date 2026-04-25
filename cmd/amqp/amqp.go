package amqp

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// amqp localized logger
var logger zerolog.Logger

// amqpCmd represents the amqp command
var amqpCmd = &cobra.Command{
	Use:   "amqp",
	Short: "Execute test commands against an AMQP server",
	Long:  "",
}

func init() {
	amqpCmd.AddCommand(amqpProduceCmd)
	amqpCmd.PersistentFlags().String("uri", "amqp://guest:guest@localhost:5672/", "AMQP URI")
	amqpCmd.PersistentFlags().String("exchange", "test-exchange", "Durable AMQP exchange name")
	amqpCmd.PersistentFlags().String("exchange-type", "direct", "Exchange type - direct|fanout|topic|x-custom")
	amqpCmd.PersistentFlags().String("queue", "test-queue", "Ephemeral AMQP queue name")
	amqpCmd.PersistentFlags().String("key", "test-key", "AMQP routing key")
	_ = viper.BindPFlag("amqp.uri", amqpCmd.PersistentFlags().Lookup("uri"))
	_ = viper.BindPFlag("amqp.exchange", amqpCmd.PersistentFlags().Lookup("exchange"))
	_ = viper.BindPFlag("amqp.exchange_type", amqpCmd.PersistentFlags().Lookup("exchange-type"))
	_ = viper.BindPFlag("amqp.queue", amqpCmd.PersistentFlags().Lookup("queue"))
	_ = viper.BindPFlag("amqp.key", amqpCmd.PersistentFlags().Lookup("key"))
}

func configLogger(cmd *cobra.Command, args []string) error {
	logger = log.Logger.With().Str("cmd", cmd.Name()).Logger()
	return nil
}
