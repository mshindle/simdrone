package rabbit

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

const Exchange = "simdrone.events"

type publishableChannel interface {
	Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
}

// amqpDispatcher is used as anchor for bus messsage method for real
// AMQP channels. For publishing to a queue, exchange should be set to the empty string
// and routingKey to the queue name.
type amqpDispatcher struct {
	channel       publishableChannel
	exchange      string
	mandatorySend bool
}

// NewAMQPDispatcher returns a new AMQP dispatcher wrapped around a single
// publishing channel.
func NewAMQPDispatcher(publishChannel publishableChannel, exchange string, mandatory bool) *amqpDispatcher {
	return &amqpDispatcher{
		channel:       publishChannel,
		exchange:      exchange,
		mandatorySend: mandatory,
	}
}

// Dispatch implementation of bus message interface method
func (a *amqpDispatcher) Dispatch(ctx context.Context, routingKey string, message any) error {
	log.Info().Fields(map[string]string{"exchange": a.exchange, "routingKey": routingKey}).Msg("message dispatching")
	body, err := json.Marshal(message)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal message")
		return err
	}

	err = a.channel.Publish(
		a.exchange,      // exchange
		routingKey,      // routing key
		a.mandatorySend, // mandatory
		false,           // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	return err
}
