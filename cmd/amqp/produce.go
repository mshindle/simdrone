package amqp

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const NumMessagesWaiting = 8
const ClientConnectionName = "producer-with-confirms"

var amqpProduceCmd = &cobra.Command{
	Use:   "produce",
	Short: "Send messages to an amqp server",
	Long: `
This example declares a durable exchange, and publishes one messages to that
exchange. This example allows up to 8 outstanding publisher confirmations
before blocking publishing. If the continuous flag is set, the example will publish
a message at a 1msg/sec rate.
`,
	PreRunE: configLogger,
	RunE:    produceMessage,
}

func init() {
	amqpProduceCmd.Flags().String("body", "foobar", "Body of message")
	amqpProduceCmd.Flags().Bool("continuous", false, "Keep publishing messages at a 1msg/sec rate")

	_ = viper.BindPFlag("amqp.body", amqpProduceCmd.Flags().Lookup("body"))
	_ = viper.BindPFlag("amqp.continuous", amqpProduceCmd.Flags().Lookup("continuous"))
}

func produceMessage(cmd *cobra.Command, args []string) error {
	exitCh := make(chan struct{})
	confirmsCh := make(chan *amqp.DeferredConfirmation)
	doneCh := make(chan struct{})
	// Note: this is a buffered channel so that indicating OK to
	// publish does not block the confirmation handler
	publishOkCh := make(chan struct{}, 1)

	setupCloseHandler(exitCh)
	startConfirmHandler(publishOkCh, confirmsCh, doneCh, exitCh)
	return publish(publishOkCh, confirmsCh, doneCh, exitCh)
}

func publish(publishOkCh <-chan struct{}, confirmsCh chan<- *amqp.DeferredConfirmation, confirmsDoneCh <-chan struct{}, exitCh chan struct{}) error {
	config := amqp.Config{
		Vhost:      "/",
		Properties: amqp.NewConnectionProperties(),
	}
	config.Properties.SetClientConnectionName(ClientConnectionName)

	uri := viper.GetString("amqp.uri")
	logger.Info().Str("uri", uri).Msg("dialing to amqp server")
	conn, err := amqp.DialConfig(uri, config)
	if err != nil {
		logger.Error().Err(err).Msg("failed to dial amqp server")
		return err
	}
	defer conn.Close()

	logger.Info().Msg("got Connection, getting Channel")
	channel, err := conn.Channel()
	if err != nil {
		logger.Error().Err(err).Msg("error getting a channel")
		return err
	}
	defer channel.Close()

	exchange := viper.GetString("amqp.exchange")
	logger.Info().Msg("declaring exchange")
	if err := channel.ExchangeDeclare(
		exchange,                              // name
		viper.GetString("amqp.exchange_type"), // type
		true,                                  // durable
		false,                                 // auto-delete
		false,                                 // internal
		false,                                 // noWait
		nil,                                   // arguments
	); err != nil {
		logger.Error().Err(err).Msg("error declaring exchange")
		return err
	}

	queueName := viper.GetString("amqp.queue")
	routingKey := viper.GetString("amqp.key")
	logger.Info().Str("queue", queueName).Msg("declaring queue")
	queue, err := channel.QueueDeclare(
		queueName, // name of the queue
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // noWait
		nil,       // arguments
	)
	if err != nil {
		logger.Error().Err(err).Str("queue", queueName).Msg("unable to declare queue")
		return err
	}

	logger.Info().Str("queue", queueName).Str("exchange", exchange).
		Int("messages", queue.Messages).Int("consumers", queue.Consumers).
		Str("routingKey", routingKey).Msg("binding queue to exchange")
	if err := channel.QueueBind(queue.Name, routingKey, exchange, false, nil); err != nil {
		logger.Error().Err(err).Msg("queue binding failed")
		return err
	}

	// Reliable publisher confirms require confirm.select support from the
	// connection.
	logger.Info().Msg("enabling publish confirms")
	if err := channel.Confirm(false); err != nil {
		logger.Error().Err(err).Msg("channel could not be put into confirm mode")
	}

	// now that we are configured, we can start publishing
	body := viper.GetString("amqp.body")
	continuous := viper.GetBool("amqp.continuous")
	for {
		canPublish := false
		logger.Info().Msg("waiting on the OK to publish...")
		for {
			select {
			case <-confirmsDoneCh:
				logger.Info().Msg("stopping, all confirms seen")
				return nil
			case <-publishOkCh:
				logger.Info().Msg("got the OK to publish")
				canPublish = true
				break
			case <-time.After(time.Second):
				logger.Warn().Msg("still waiting on the OK to publish...")
				continue
			}
			if canPublish {
				break
			}
		}

		logger.Info().Int("bytes", len(body)).Str("body", body).Msg("publishing body")
		dConfirmation, err := channel.PublishWithDeferredConfirm(
			exchange,
			routingKey,
			true,
			false,
			amqp.Publishing{
				Headers:         amqp.Table{},
				ContentType:     "text/plain",
				ContentEncoding: "",
				DeliveryMode:    amqp.Persistent,
				Priority:        0,
				AppId:           "sequential-producer",
				Body:            []byte(body),
			},
		)
		if err != nil {
			logger.Error().Err(err).Msg("error in publish")
		}

		select {
		case <-confirmsDoneCh:
			logger.Info().Msg("stopping, all confirms seen")
			return nil
		case confirmsCh <- dConfirmation:
			logger.Info().Msg("delivered deferred confirmation")
			break
		}

		select {
		case <-confirmsDoneCh:
			logger.Info().Msg("stopping, all confirms seen")
			return nil
		case <-time.After(250 * time.Millisecond):
			if continuous {
				continue
			}
			logger.Info().Msg("initiating stop")
			close(exitCh)
			select {
			case <-confirmsDoneCh:
				logger.Info().Msg("stopping, all confirms seen")
				break
			case <-time.After(time.Second * 10):
				logger.Warn().Msg("may be stopping with outstanding confirmations")
				break
			}
			break
		}
		break
	}
	return nil
}

func setupCloseHandler(exitCh chan struct{}) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Logger.Warn().Msg("Received interrupt, exiting...")
		close(exitCh)
	}()
}

func startConfirmHandler(publishOkCh chan<- struct{}, confirmsCh <-chan *amqp.DeferredConfirmation, confirmsDoneCh chan struct{}, exitCh <-chan struct{}) {
	go func() {
		confirms := make(map[uint64]*amqp.DeferredConfirmation)

		for {
			select {
			case <-exitCh:
				exitConfirmHandler(confirms, confirmsDoneCh)
				return
			default:
				break
			}

			outstandingConfirmationCount := len(confirms)

			// Note: 8 is arbitrary, you may wish to allow more outstanding confirms before blocking publish
			if outstandingConfirmationCount <= NumMessagesWaiting {
				select {
				case publishOkCh <- struct{}{}:
					log.Logger.Info().Msg("confirm handler: sent OK to publish")
				case <-time.After(time.Second * 5):
					log.Logger.Warn().Msg("confirm handler: timeout indicating OK to publish (this should never happen!)")
				}
			} else {
				log.Logger.Warn().
					Int("confirms", outstandingConfirmationCount).
					Msg("confirm handler: waiting on outstanding confirmations, blocking publish")
			}

			select {
			case confirmation := <-confirmsCh:
				dtag := confirmation.DeliveryTag
				confirms[dtag] = confirmation
			case <-exitCh:
				exitConfirmHandler(confirms, confirmsDoneCh)
				return
			}

			checkConfirmations(confirms)
		}
	}()
}

func exitConfirmHandler(confirms map[uint64]*amqp.DeferredConfirmation, confirmsDoneCh chan struct{}) {
	log.Logger.Info().Msg("confirm handler: exit requested")
	waitConfirmations(confirms)
	close(confirmsDoneCh)
	log.Logger.Info().Msg("confirm handler: exiting")
}

func waitConfirmations(confirms map[uint64]*amqp.DeferredConfirmation) {
	log.Logger.Info().Int("confirms", len(confirms)).Msg("confirm handler: waiting on outstanding confirmations")
	checkConfirmations(confirms)

	for k, v := range confirms {
		select {
		case <-v.Done():
			log.Logger.Info().Uint64("tag", k).Msg("confirm handler: confirmed delivery with tag")
			delete(confirms, k)
		case <-time.After(time.Second):
			log.Logger.Warn().Uint64("tag", k).Msg("confirm handler: did not receive confirmation for tag")
		}
	}

	outstandingConfirmationCount := len(confirms)
	if outstandingConfirmationCount > 0 {
		log.Logger.Error().Int("confirms", outstandingConfirmationCount).
			Msg("confirm handler: exiting with outstanding confirmations")
	} else {
		log.Logger.Info().Msg("confirm handler: done waiting on outstanding confirmations")
	}
}

func checkConfirmations(confirms map[uint64]*amqp.DeferredConfirmation) {
	log.Logger.Info().Int("confirms", len(confirms)).Msg("confirm handler: checking outstanding confirmations")
	for k, v := range confirms {
		if v.Acked() {
			log.Logger.Info().Uint64("tag", k).Msg("confirm handler: confirmed delivery with tag")
			delete(confirms, k)
		}
	}
}
