package bus

import "time"

// Exchange is the name of the topic for submitting events
const (
	Exchange        = "simdrone.events"
	TelemetryUpdate = "telemetry.updated"
	PositionUpdate  = "position.updated"
	AlertSignal     = "alert.signalled"
)

type Dispatcher interface {
	Dispatch(routingKey string, message interface{}) error
}

type Config struct {
	URL string `json:"url"`
}

// TelemetryUpdatedEvent is an event containing telemetry updates received from a drone
type TelemetryUpdatedEvent struct {
	DroneID          string    `json:"drone_id"`
	RemainingBattery int       `json:"battery"`
	Uptime           int       `json:"uptime"`
	CoreTemp         int       `json:"core_temp"`
	ReceivedOn       time.Time `json:"received_on"`
}

// AlertSignalledEvent is an event indicating an alert condition was reported by a drone
type AlertSignalledEvent struct {
	DroneID     string    `json:"drone_id"`
	FaultCode   int       `json:"fault_code"`
	Description string    `json:"description"`
	ReceivedOn  time.Time `json:"received_on"`
}

// PositionChangedEvent is an event indicating that the position and speed of a drone was reported.
type PositionChangedEvent struct {
	DroneID         string    `json:"drone_id"`
	Latitude        float32   `json:"latitude"`
	Longitude       float32   `json:"longitude"`
	Altitude        float32   `json:"altitude"`
	CurrentSpeed    float32   `json:"current_speed"`
	HeadingCardinal int       `json:"heading_cardinal"`
	ReceivedOn      time.Time `json:"received_on"`
}
