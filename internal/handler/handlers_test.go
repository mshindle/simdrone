package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/mshindle/simdrone/internal/event"
)

type mockDispatcher struct {
	mu       sync.Mutex
	Messages []any
}

func (m *mockDispatcher) Dispatch(_ context.Context, _ string, message any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, message)
	return nil
}

func TestAddValidTelemetryCreatesCommand(t *testing.T) {
	dispatcher := &mockDispatcher{}
	h := New(dispatcher)

	body := []byte(`{"drone_id":"drone666", "battery": 72, "uptime": 6941, "core_temp": 21 }`)
	req := httptest.NewRequest(http.MethodPost, "/api/cmds/telemetry", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.e.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected 201, got %d. Body: %s", rec.Code, rec.Body.String())
	}
	if len(dispatcher.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(dispatcher.Messages))
	}

	var telemetryResponse event.TelemetryUpdated
	if err := json.Unmarshal(rec.Body.Bytes(), &telemetryResponse); err != nil {
		t.Fatalf("Could not unmarshal response: %v", err)
	}
	if telemetryResponse.DroneID != "drone666" {
		t.Errorf("Expected drone666, got %s", telemetryResponse.DroneID)
	}
}

func TestAddInvalidTelemetryReturnsBadRequest(t *testing.T) {
	dispatcher := &mockDispatcher{}
	h := New(dispatcher)

	body := []byte(`{"foo":"bar"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/cmds/telemetry", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestAddValidPositionCreatesCommand(t *testing.T) {
	dispatcher := &mockDispatcher{}
	h := New(dispatcher)

	body := []byte(`{"drone_id":"pos1", "latitude": 45.0, "longitude": 90.0, "altitude": 100.0, "current_speed": 10.0, "heading_cardinal": 180}`)
	req := httptest.NewRequest(http.MethodPost, "/api/cmds/position", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.e.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected 201, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	var posResponse event.PositionChanged
	if err := json.Unmarshal(rec.Body.Bytes(), &posResponse); err != nil {
		t.Fatalf("Could not unmarshal response: %v", err)
	}
	if posResponse.DroneID != "pos1" {
		t.Errorf("Expected pos1, got %s", posResponse.DroneID)
	}
}

func TestAddInvalidPositionCommandReturnsBadRequest(t *testing.T) {
	dispatcher := &mockDispatcher{}
	h := New(dispatcher)

	body := []byte(`{"drone_id":"pos1", "latitude": 1000.0}`) // Invalid latitude
	req := httptest.NewRequest(http.MethodPost, "/api/cmds/position", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}

func TestAddValidAlertCreatesCommand(t *testing.T) {
	dispatcher := &mockDispatcher{}
	h := New(dispatcher)

	body := []byte(`{"drone_id":"drone1", "fault_code": 1, "description": "Engine failure"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/cmds/alert", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.e.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected 201, got %d", rec.Code)
	}

	var alertResponse event.AlertSignalled
	if err := json.Unmarshal(rec.Body.Bytes(), &alertResponse); err != nil {
		t.Fatalf("Could not unmarshal response: %v", err)
	}
	if alertResponse.Description != "Engine failure" {
		t.Errorf("Expected 'Engine failure', got '%s'", alertResponse.Description)
	}
}

func TestAddInvalidAlertCommandReturnsBadRequest(t *testing.T) {
	dispatcher := &mockDispatcher{}
	h := New(dispatcher)

	body := []byte(`{"fault_code": 1}`) // Missing drone_id and description
	req := httptest.NewRequest(http.MethodPost, "/api/cmds/alert", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", rec.Code)
	}
}
