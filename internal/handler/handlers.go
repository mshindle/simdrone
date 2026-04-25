package handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/mshindle/simdrone/internal/bus"
	"github.com/mshindle/simdrone/internal/event"
	"github.com/mshindle/simdrone/internal/web"
)

func bindCommand(c *echo.Context, cmd any) error {
	// 1. Bind: Map JSON/Form data to the struct
	if err := c.Bind(cmd); err != nil {
		return err
	}

	// 2. Validate: Run the rules defined in the struct tags
	if err := c.Validate(cmd); err != nil {
		return err // Returns the 400 Bad Request from our CustomValidator
	}
	return nil
}

func dispatchCommand(c *echo.Context, dispatcher bus.Dispatcher, key string, evt any) error {
	// and send it...
	l := web.LoggerFromEchoContext(c)
	err := dispatcher.Dispatch(c.Request().Context(), key, evt)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to dispatch alert signal").Wrap(err)
	}
	l.Debug().Msg("dispatched event")
	return c.JSON(http.StatusCreated, evt)
}

func (h *Handler) alertHandler(c *echo.Context) error {
	l := web.LoggerFromEchoContext(c)
	cmd := new(alertCommand)
	err := bindCommand(c, cmd)
	if err != nil {
		return err
	}
	l.Debug().Str("drone_id", cmd.DroneID).Msg("received cmd")

	// convert to an event
	evt := &event.AlertSignalled{
		DroneID:     cmd.DroneID,
		FaultCode:   cmd.FaultCode,
		Description: cmd.Description,
		ReceivedAt:  time.Now(),
	}

	return dispatchCommand(c, h.dispatcher, event.AlertSignal, evt)
}

func (h *Handler) telemetryHandler(c *echo.Context) error {
	cmd := new(telemetryCommand)
	err := bindCommand(c, cmd)
	if err != nil {
		return err
	}
	evt := &event.TelemetryUpdated{
		DroneID:          cmd.DroneID,
		RemainingBattery: cmd.RemainingBattery,
		Uptime:           cmd.Uptime,
		CoreTemp:         cmd.CoreTemp,
		ReceivedAt:       time.Now(),
	}
	return dispatchCommand(c, h.dispatcher, event.TelemetryUpdate, evt)
}

func (h *Handler) positionHandler(c *echo.Context) error {
	cmd := new(positionCommand)
	err := bindCommand(c, cmd)
	if err != nil {
		return err
	}
	evt := &event.PositionChanged{
		DroneID:         cmd.DroneID,
		Longitude:       cmd.Longitude,
		Latitude:        cmd.Latitude,
		Altitude:        cmd.Altitude,
		CurrentSpeed:    cmd.CurrentSpeed,
		HeadingCardinal: cmd.HeadingCardinal,
		ReceivedAt:      time.Now(),
	}
	return dispatchCommand(c, h.dispatcher, event.PositionUpdate, evt)
}
