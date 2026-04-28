package mongodb

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type mongoTelemetryRecord struct {
	RecordID         bson.ObjectID `bson:"_id omitempty" json:"id"`
	DroneID          string        `bson:"drone_id" json:"drone_id"`
	RemainingBattery int           `bson:"remaining_battery" json:"remaining_battery"`
	Uptime           int           `bson:"uptime" json:"uptime"`
	CoreTemp         int           `bson:"core_temp" json:"core_temp"`
	ReceivedAt       time.Time     `bson:"received_at" json:"received_at"`
}

type mongoAlertRecord struct {
	RecordID    bson.ObjectID `bson:"_id omitempty" json:"id"`
	DroneID     string        `bson:"drone_id" json:"drone_id"`
	FaultCode   int           `bson:"fault_code" json:"fault_code"`
	Description string        `bson:"description" json:"description"`
	ReceivedAt  time.Time     `bson:"received_at" json:"received_at"`
}

type mongoPositionRecord struct {
	RecordID        bson.ObjectID `bson:"_id omitempty" json:"id"`
	DroneID         string        `bson:"drone_id" json:"drone_id"`
	Latitude        float32       `bson:"latitutde" json:"latitude"`
	Longitude       float32       `bson:"longitude" json:"longitude"`
	Altitude        float32       `bson:"altitude" json:"altitude"`
	CurrentSpeed    float32       `bson:"current_speed" json:"current_speed"`
	HeadingCardinal int           `bson:"heading_cardinal" json:"heading_cardinal"`
	ReceivedAt      time.Time     `bson:"received_at" json:"received_at"`
}

type mongoRecord interface {
	mongoTelemetryRecord | mongoAlertRecord | mongoPositionRecord
}
