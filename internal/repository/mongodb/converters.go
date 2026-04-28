package mongodb

import (
	"unicode"

	"github.com/mshindle/simdrone/internal/event"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func convertAlertEventToRecord(evt *event.AlertSignalled, recordID bson.ObjectID) *mongoAlertRecord {
	return &mongoAlertRecord{
		RecordID:    recordID,
		DroneID:     evt.DroneID,
		FaultCode:   evt.FaultCode,
		Description: evt.Description,
		ReceivedAt:  evt.ReceivedAt,
	}
}

func convertTelemetryEventToRecord(evt *event.TelemetryUpdated, recordID bson.ObjectID) *mongoTelemetryRecord {
	return &mongoTelemetryRecord{
		RecordID:         recordID,
		DroneID:          evt.DroneID,
		RemainingBattery: evt.RemainingBattery,
		Uptime:           evt.Uptime,
		CoreTemp:         evt.CoreTemp,
		ReceivedAt:       evt.ReceivedAt,
	}
}

func convertPositionEventToRecord(evt *event.PositionChanged, recordID bson.ObjectID) *mongoPositionRecord {
	return &mongoPositionRecord{
		RecordID:        recordID,
		DroneID:         evt.DroneID,
		Latitude:        evt.Latitude,
		Longitude:       evt.Longitude,
		Altitude:        evt.Altitude,
		CurrentSpeed:    evt.CurrentSpeed,
		HeadingCardinal: evt.HeadingCardinal,
		ReceivedAt:      evt.ReceivedAt,
	}
}

func convertTelemetryRecordToEvent(record *mongoTelemetryRecord) *event.TelemetryUpdated {
	return &event.TelemetryUpdated{
		DroneID:          record.DroneID,
		RemainingBattery: record.RemainingBattery,
		Uptime:           record.Uptime,
		CoreTemp:         record.CoreTemp,
		ReceivedAt:       record.ReceivedAt,
	}
}

func convertPositionRecordToEvent(record *mongoPositionRecord) *event.PositionChanged {
	return &event.PositionChanged{
		DroneID:         record.DroneID,
		Altitude:        record.Altitude,
		CurrentSpeed:    record.CurrentSpeed,
		HeadingCardinal: record.HeadingCardinal,
		Latitude:        record.Latitude,
		Longitude:       record.Longitude,
		ReceivedAt:      record.ReceivedAt,
	}
}

func convertAlertRecordToEvent(record *mongoAlertRecord) *event.AlertSignalled {
	return &event.AlertSignalled{
		DroneID:     record.DroneID,
		Description: record.Description,
		FaultCode:   record.FaultCode,
		ReceivedAt:  record.ReceivedAt,
	}
}

func upperFirst(s string) string {
	if s == "" {
		return ""
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}
