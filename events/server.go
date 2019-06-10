package events

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mshindle/simdrone/bus"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
)

// Config holds the real world values used to manage events
type Config struct {
	DispatchConfig     bus.Config
	AlertsQueueName    string
	TelemetryQueueName string
	PositionsQueueName string
}

func NewServer(config *Config) *negroni.Negroni {
	// create our processors
	alertProc := alertProcessor{config.AlertsQueueName, make(chan bus.AlertSignalledEvent)}
	telemetryProc := telemetryProcessor{config.TelemetryQueueName, make(chan bus.TelemetryUpdatedEvent)}
	positionProc := positionProcessor{config.PositionsQueueName, make(chan bus.PositionChangedEvent)}

	// initialize our queue readers
	err := initQueueReaders(config, alertProc, telemetryProc, positionProc)
	if err != nil {
		logrus.WithError(err).Error("no dispatcher available")
	}

	// initialize our data repository
	repo :=
	// configure simple http server
	n := negroni.Classic()
	router := mux.NewRouter()
	initRoutes(router)
	n.UseHandler(router)

	return n
}

func initQueueReaders(config *Config, processors ...Processor) error {
	// create an amqp dispatcher
	logrus.WithField("url", config.DispatchConfig.URL).Info("connecting to amqp server")
	conn, err := amqp.Dial(config.DispatchConfig.URL)
	if err != nil {
		return err
	}

	channel, err := conn.Channel()
	if err != nil {
		return err
	}

	for _, proc := range processors {
		queue, err := channel.QueueDeclare(
			proc.QueueName(),
			false,
			false,
			false,
			false,
			nil)
		if err != nil {
			return err
		}

		msgIn, err := channel.Consume(
			queue.Name,
			"",
			true,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return err
		}

		go func() {
			for {
				select {
				case msgRaw := <-msgIn:
					proc.Dispatch(msgRaw)
				}
			}
		}()
	}
}

func initRoutes(router *mux.Router) {
	formatter := render.New(render.Options{IndentJSON: true})
	router.HandleFunc("/", homeHandler(formatter)).Methods("GET")
}

func homeHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		formatter.Text(w, http.StatusOK, "Simdrone- Event Processor see http://github.com/mshindle/simdrone")
	}
}
