package mongo

import (
	"time"

	"github.com/mshindle/simdrone/bus"
	"gopkg.in/mgo.v2/bson"
)

func convertAlertEventToRecord(event bus.AlertSignalledEvent, recordID bson.ObjectId) (record *mongoAlertRecord) {
	record = &mongoAlertRecord{
		RecordID:    recordID,
		DroneID:     event.DroneID,
		FaultCode:   event.FaultCode,
		Description: event.Description,
		ReceivedOn:  event.ReceivedOn.Format(time.UnixDate),
	}
	return
}

func convertTelemetryEventToRecord(event bus.TelemetryUpdatedEvent, recordID bson.ObjectId) (record *mongoTelemetryRecord) {
	record = &mongoTelemetryRecord{
		RecordID:         recordID,
		DroneID:          event.DroneID,
		RemainingBattery: event.RemainingBattery,
		Uptime:           event.Uptime,
		CoreTemp:         event.CoreTemp,
		ReceivedOn:       event.ReceivedOn.Format(time.UnixDate),
	}

	return
}

func convertPositionEventToRecord(event bus.PositionChangedEvent, recordID bson.ObjectId) (record *mongoPositionRecord) {
	record = &mongoPositionRecord{
		RecordID:        recordID,
		DroneID:         event.DroneID,
		Latitude:        event.Latitude,
		Longitude:       event.Longitude,
		Altitude:        event.Altitude,
		CurrentSpeed:    event.CurrentSpeed,
		HeadingCardinal: event.HeadingCardinal,
		ReceivedOn:      event.ReceivedOn.Format(time.UnixDate),
	}
	return
}

func convertTelemetryRecordToEvent(record mongoTelemetryRecord) (event bus.TelemetryUpdatedEvent) {
	t, _ := time.Parse(time.UnixDate, record.ReceivedOn)
	event = bus.TelemetryUpdatedEvent{
		DroneID:          record.DroneID,
		RemainingBattery: record.RemainingBattery,
		Uptime:           record.Uptime,
		CoreTemp:         record.CoreTemp,
		ReceivedOn:       t,
	}
	return
}

func convertPositionRecordToEvent(record mongoPositionRecord) (event bus.PositionChangedEvent) {
	t, _ := time.Parse(time.UnixDate, record.ReceivedOn)
	event = bus.PositionChangedEvent{
		DroneID:         record.DroneID,
		Altitude:        record.Altitude,
		CurrentSpeed:    record.CurrentSpeed,
		HeadingCardinal: record.HeadingCardinal,
		Latitude:        record.Latitude,
		Longitude:       record.Longitude,
		ReceivedOn:      t,
	}
	return
}

func convertAlertRecordToEvent(record mongoAlertRecord) (event bus.AlertSignalledEvent) {
	t, _ := time.Parse(time.UnixDate, record.ReceivedOn)
	event = bus.AlertSignalledEvent{
		DroneID:     record.DroneID,
		Description: record.Description,
		FaultCode:   record.FaultCode,
		ReceivedOn:  t,
	}
	return
}
