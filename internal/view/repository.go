package view

import (
	"context"
	"fmt"
	"time"

	"github.com/mshindle/simdrone/internal/event"
)

type EventRepository interface {
	GetAlert(ctx context.Context, droneID string) (*event.AlertSignalled, error)
	GetTelemetry(ctx context.Context, droneID string) (*event.TelemetryUpdated, error)
	GetPosition(ctx context.Context, droneID string) (*event.PositionChanged, error)
	GetActiveDrones(ctx context.Context, d time.Duration) ([]string, error)
}

type mockEventRepository struct{}

func (m *mockEventRepository) GetAlert(_ context.Context, droneID string) (*event.AlertSignalled, error) {
	return nil, fmt.Errorf("not implemented for drones [%s]", droneID)
}

func (m *mockEventRepository) GetTelemetry(_ context.Context, droneID string) (*event.TelemetryUpdated, error) {
	return nil, fmt.Errorf("not implemented for drones [%s]", droneID)
}

func (m *mockEventRepository) GetPosition(_ context.Context, droneID string) (*event.PositionChanged, error) {
	return nil, fmt.Errorf("not implemented for drones [%s]", droneID)
}

func (m *mockEventRepository) GetActiveDrones(_ context.Context, _ time.Duration) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}
