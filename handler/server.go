package handler

import (
	"strings"

	"github.com/gorilla/mux"
	"github.com/mshindle/simdrone/bus"
	"github.com/mshindle/simdrone/bus/dispatch"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
)

type Config struct {
	DispatchConfig bus.Config
}

func NewServer(config *Config) *negroni.Negroni {
	// create our dispatcher
	dispatcher, err := createDispatcher(config.DispatchConfig)
	if err != nil {
		logrus.WithError(err).Error("no dispatcher available")
	}

	n := negroni.Classic()
	router := mux.NewRouter()
	initRoutes(router, dispatcher)
	n.UseHandler(router)

	return n
}

func initRoutes(router *mux.Router, dispatcher bus.Dispatcher) {
	formatter := render.New(render.Options{IndentJSON: true})
	router.HandleFunc("/api/cmds/alert", addAlertHandler(formatter, dispatcher)).Methods("POST")
	router.HandleFunc("/api/cmds/telemetry", addTelemetryHandler(formatter, dispatcher)).Methods("POST")
	router.HandleFunc("/api/cmds/position", addPositionHandler(formatter, dispatcher)).Methods("POST")
}

func createDispatcher(config bus.Config) (bus.Dispatcher, error) {
	if strings.HasPrefix(config.URL, "fake://") {
		logrus.Info("creating mock dispatcher")
		return dispatch.NewMockDispatcher(), nil
	}
	// create an amqp dispatcher
	logrus.WithField("url", config.URL).Info("connecting to amqp server")
	conn, err := amqp.Dial(config.URL)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	err = channel.ExchangeDeclare(
		bus.Exchange, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		return nil, err
	}

	return dispatch.NewAMQPDispatcher(channel, bus.Exchange, false), nil
}
