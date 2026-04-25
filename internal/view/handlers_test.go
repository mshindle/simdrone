package view

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/mshindle/simdrone/internal/event"
	"github.com/mshindle/simdrone/internal/repository"
)

type mockRepo struct {
	GetAlertFunc        func(ctx context.Context, droneID string) (*event.AlertSignalled, error)
	GetTelemetryFunc    func(ctx context.Context, droneID string) (*event.TelemetryUpdated, error)
	GetPositionFunc     func(ctx context.Context, droneID string) (*event.PositionChanged, error)
	GetActiveDronesFunc func(ctx context.Context, d time.Duration) ([]string, error)
}

func (m *mockRepo) GetAlert(ctx context.Context, droneID string) (*event.AlertSignalled, error) {
	if m.GetAlertFunc != nil {
		return m.GetAlertFunc(ctx, droneID)
	}
	return nil, nil
}

func (m *mockRepo) GetTelemetry(ctx context.Context, droneID string) (*event.TelemetryUpdated, error) {
	if m.GetTelemetryFunc != nil {
		return m.GetTelemetryFunc(ctx, droneID)
	}
	return nil, nil
}

func (m *mockRepo) GetPosition(ctx context.Context, droneID string) (*event.PositionChanged, error) {
	if m.GetPositionFunc != nil {
		return m.GetPositionFunc(ctx, droneID)
	}
	return nil, nil
}

func (m *mockRepo) GetActiveDrones(ctx context.Context, d time.Duration) ([]string, error) {
	if m.GetActiveDronesFunc != nil {
		return m.GetActiveDronesFunc(ctx, d)
	}
	return nil, nil
}

func TestActiveHandler(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		repoResult     []string
		repoErr        error
		expectedStatus int
		expectedBody   any
	}{
		{
			name:           "SuccessDefaultDuration",
			query:          "",
			repoResult:     []string{"drone1", "drone2"},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"drone1", "drone2"},
		},
		{
			name:           "SuccessExplicitDuration",
			query:          "?d=10m",
			repoResult:     []string{"drone3"},
			expectedStatus: http.StatusOK,
			expectedBody:   []string{"drone3"},
		},
		{
			name:           "InvalidDuration",
			query:          "?d=invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "RepoError",
			repoErr:        errors.New("db error"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "NoActiveDrones",
			repoErr:        repository.ErrNotFound,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{
				GetActiveDronesFunc: func(ctx context.Context, d time.Duration) ([]string, error) {
					if tt.name == "SuccessExplicitDuration" && d != 10*time.Minute {
						t.Errorf("expected duration 10m, got %v", d)
					}
					if tt.name == "SuccessDefaultDuration" && d != 5*time.Minute {
						t.Errorf("expected default duration 5m, got %v", d)
					}
					return tt.repoResult, tt.repoErr
				},
			}
			srv := New(WithRepository(repo))

			req := httptest.NewRequest(http.MethodGet, "/drones"+tt.query, nil)
			rec := httptest.NewRecorder()
			srv.e.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d. Body: %s", tt.expectedStatus, rec.Code, rec.Body.String())
			}

			if tt.expectedBody != nil {
				var gotBody []string
				if err := json.Unmarshal(rec.Body.Bytes(), &gotBody); err != nil {
					t.Fatalf("failed to unmarshal body: %v", err)
				}
				if !reflect.DeepEqual(gotBody, tt.expectedBody) {
					t.Errorf("expected body %v, got %v", tt.expectedBody, gotBody)
				}
			}
		})
	}
}

func TestLastEventHandler(t *testing.T) {
	t.Run("AlertSuccess", func(t *testing.T) {
		expected := &event.AlertSignalled{DroneID: "drone1", Description: "Fire"}
		repo := &mockRepo{
			GetAlertFunc: func(ctx context.Context, droneID string) (*event.AlertSignalled, error) {
				if droneID != "drone1" {
					return nil, fmt.Errorf("expected droneID drone1, got %s", droneID)
				}
				return expected, nil
			},
		}
		srv := New(WithRepository(repo))

		req := httptest.NewRequest(http.MethodGet, "/drones/drone1/lastAlert", nil)
		rec := httptest.NewRecorder()
		srv.e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var got event.AlertSignalled
		json.Unmarshal(rec.Body.Bytes(), &got)
		if got.DroneID != expected.DroneID || got.Description != expected.Description {
			t.Errorf("expected %v, got %v", expected, got)
		}
	})

	t.Run("TelemetrySuccess", func(t *testing.T) {
		expected := &event.TelemetryUpdated{DroneID: "drone1", RemainingBattery: 50}
		repo := &mockRepo{
			GetTelemetryFunc: func(ctx context.Context, droneID string) (*event.TelemetryUpdated, error) {
				if droneID != "drone1" {
					return nil, fmt.Errorf("expected droneID drone1, got %s", droneID)
				}
				return expected, nil
			},
		}
		srv := New(WithRepository(repo))

		req := httptest.NewRequest(http.MethodGet, "/drones/drone1/lastTelemetry", nil)
		rec := httptest.NewRecorder()
		srv.e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var got event.TelemetryUpdated
		json.Unmarshal(rec.Body.Bytes(), &got)
		if got.DroneID != expected.DroneID || got.RemainingBattery != expected.RemainingBattery {
			t.Errorf("expected %v, got %v", expected, got)
		}
	})

	t.Run("PositionSuccess", func(t *testing.T) {
		expected := &event.PositionChanged{DroneID: "drone1", Altitude: 100}
		repo := &mockRepo{
			GetPositionFunc: func(ctx context.Context, droneID string) (*event.PositionChanged, error) {
				if droneID != "drone1" {
					return nil, fmt.Errorf("expected droneID drone1, got %s", droneID)
				}
				return expected, nil
			},
		}
		srv := New(WithRepository(repo))

		req := httptest.NewRequest(http.MethodGet, "/drones/drone1/lastPosition", nil)
		rec := httptest.NewRecorder()
		srv.e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
		}

		var got event.PositionChanged
		json.Unmarshal(rec.Body.Bytes(), &got)
		if got.DroneID != expected.DroneID || got.Altitude != expected.Altitude {
			t.Errorf("expected %v, got %v", expected, got)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		repo := &mockRepo{
			GetAlertFunc: func(ctx context.Context, droneID string) (*event.AlertSignalled, error) {
				return nil, repository.ErrNotFound
			},
		}
		srv := New(WithRepository(repo))

		req := httptest.NewRequest(http.MethodGet, "/drones/drone1/lastAlert", nil)
		rec := httptest.NewRecorder()
		srv.e.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})

	t.Run("InternalError", func(t *testing.T) {
		repo := &mockRepo{
			GetAlertFunc: func(ctx context.Context, droneID string) (*event.AlertSignalled, error) {
				return nil, errors.New("db error")
			},
		}
		srv := New(WithRepository(repo))

		req := httptest.NewRequest(http.MethodGet, "/drones/drone1/lastAlert", nil)
		rec := httptest.NewRecorder()
		srv.e.ServeHTTP(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", rec.Code)
		}
	})
}
