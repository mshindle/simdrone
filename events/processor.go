package events

import (
	"encoding/json"

	"github.com/mshindle/simdrone/bus"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type Processor interface {
	QueueName() string
	Dispatch(delivery amqp.Delivery)
}

type alertProcessor struct {
	queue   string
	channel chan bus.AlertSignalledEvent
}

func (a alertProcessor) QueueName() string {
	return a.queue
}

func (a alertProcessor) Dispatch(delivery amqp.Delivery) {
	var event bus.AlertSignalledEvent
	err := json.Unmarshal(delivery.Body, &event)
	if err != nil {
		logrus.WithError(err).Error("failed to deserialize alert from raw")
		return
	}
	a.channel <- event
}

type telemetryProcessor struct {
	queue   string
	channel chan bus.TelemetryUpdatedEvent
}

func (t telemetryProcessor) QueueName() string {
	return t.queue
}

func (t telemetryProcessor) Dispatch(delivery amqp.Delivery) {
	var event bus.TelemetryUpdatedEvent
	err := json.Unmarshal(delivery.Body, &event)
	if err != nil {
		logrus.WithError(err).Error("failed to deserialize telemetry from raw")
		return
	}
	t.channel <- event
}

type positionProcessor struct {
	queue   string
	channel chan bus.PositionChangedEvent
}

func (p positionProcessor) QueueName() string {
	return p.queue
}

func (p positionProcessor) Dispatch(delivery amqp.Delivery) {
	var event bus.PositionChangedEvent
	err := json.Unmarshal(delivery.Body, &event)
	if err != nil {
		logrus.WithError(err).Error("failed to deserialize position from raw")
		return
	}
	p.channel <- event
}
