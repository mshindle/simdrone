package handler

type alertCommand struct {
	DroneID     string `json:"drone_id" validate:"required"`
	FaultCode   int    `json:"fault_code"`
	Description string `json:"description" validate:"required"`
}

type telemetryCommand struct {
	DroneID          string `json:"drone_id" validate:"required"`
	RemainingBattery int    `json:"battery"`
	Uptime           int    `json:"uptime" validate:"required,gt=0"`
	CoreTemp         int    `json:"core_temp"`
}

type positionCommand struct {
	DroneID         string  `json:"drone_id" validate:"required"`
	Latitude        float32 `json:"latitude" validate:"required,latitude"`
	Longitude       float32 `json:"longitude" validate:"required,longitude"`
	Altitude        float32 `json:"altitude"`
	CurrentSpeed    float32 `json:"current_speed"`
	HeadingCardinal int     `json:"heading_cardinal" validate:"required,gte=0,lte=359"`
}
