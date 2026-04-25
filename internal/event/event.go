package event

import "time"

const (
	TelemetryUpdate string = "events.drone.telemetry.updated"
	PositionUpdate  string = "events.drone.position.updated"
	AlertSignal     string = "events.drone.alert.signalled"
	Ping            string = "events.system.ping"
)

// TelemetryUpdated is an event containing telemetry updates received from a drone
type TelemetryUpdated struct {
	DroneID          string    `json:"drone_id"`
	RemainingBattery int       `json:"battery"`
	Uptime           int       `json:"uptime"`
	CoreTemp         int       `json:"core_temp"`
	ReceivedAt       time.Time `json:"received_at"`
}

// AlertSignalled is an event indicating a drone reported an alert condition
type AlertSignalled struct {
	DroneID     string    `json:"drone_id"`
	FaultCode   int       `json:"fault_code"`
	Description string    `json:"description"`
	ReceivedAt  time.Time `json:"received_at"`
}

// PositionChanged is an event indicating that the position and speed of a drone was reported.
type PositionChanged struct {
	DroneID         string    `json:"drone_id"`
	Latitude        float32   `json:"latitude"`
	Longitude       float32   `json:"longitude"`
	Altitude        float32   `json:"altitude"`
	CurrentSpeed    float32   `json:"current_speed"`
	HeadingCardinal int       `json:"heading_cardinal"`
	ReceivedAt      time.Time `json:"received_at"`
}

// Pinged is an event indicating that the server is alive and receiving messages
type Pinged struct {
	Message    string `json:"message"`
	ReceivedAt time.Time
}

// DroneEvents is a union of all events related to a drone
type DroneEvents interface {
	TelemetryUpdated | PositionChanged | AlertSignalled
}
