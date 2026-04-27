package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/mshindle/simdrone/internal/event"
	"github.com/mshindle/simdrone/internal/web"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

const (
	attrSubject           = "bus.subject"
	attrDroneID           = "drone.id"
	attrAlertFaultCode    = "alert.fault_code"
	attrTelemetryBattery  = "telemetry.battery"
	attrTelemetryCoreTemp = "telemetry.core_temp"
	attrPositionAltitude  = "position.altitude"
	attrPositionSpeed     = "position.speed"
	errBindFailed         = "bind failed"
	errDispatchFailed     = "dispatch failed"
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

func (h *Handler) dispatchCommand(ctx context.Context, subject string, evt any) error {
	ctx, span := h.tp.Tracer(serverName).Start(ctx, "DispatchCommand")
	defer span.End()
	span.SetAttributes(attribute.String(attrSubject, subject))

	l := web.LoggerFromContext(ctx)
	err := h.dispatcher.Dispatch(ctx, subject, evt)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, errDispatchFailed)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to dispatch alert signal").Wrap(err)
	}
	l.Debug().Msg("dispatched event")
	return nil
}

func (h *Handler) alertHandler(c *echo.Context) error {
	ctx, span := h.tp.Tracer(serverName).Start(c.Request().Context(), "Handler.Alert")
	defer span.End()

	l := web.LoggerFromContext(ctx)
	cmd := new(alertCommand)
	err := bindCommand(c, cmd)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, errBindFailed)
		return err
	}

	// Tag the span with business-critical metadata
	span.SetAttributes(
		attribute.String(attrDroneID, cmd.DroneID),
		attribute.Int(attrAlertFaultCode, cmd.FaultCode),
	)
	l.Debug().Str("drone_id", cmd.DroneID).Msg("received cmd")

	// convert to an event
	evt := &event.AlertSignalled{
		DroneID:     cmd.DroneID,
		FaultCode:   cmd.FaultCode,
		Description: cmd.Description,
		ReceivedAt:  time.Now(),
	}

	err = h.dispatchCommand(ctx, event.AlertSignal, evt)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, evt)
}

func (h *Handler) telemetryHandler(c *echo.Context) error {
	ctx, span := h.tp.Tracer(serverName).Start(c.Request().Context(), "Handler.Telemetry")
	defer span.End()

	cmd := new(telemetryCommand)
	err := bindCommand(c, cmd)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, errBindFailed)
		return err
	}

	span.SetAttributes(
		attribute.String(attrDroneID, cmd.DroneID),
		attribute.Int(attrTelemetryBattery, cmd.RemainingBattery),
		attribute.Int(attrTelemetryCoreTemp, cmd.CoreTemp),
	)
	evt := &event.TelemetryUpdated{
		DroneID:          cmd.DroneID,
		RemainingBattery: cmd.RemainingBattery,
		Uptime:           cmd.Uptime,
		CoreTemp:         cmd.CoreTemp,
		ReceivedAt:       time.Now(),
	}

	err = h.dispatchCommand(ctx, event.TelemetryUpdate, evt)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, evt)
}

func (h *Handler) positionHandler(c *echo.Context) error {
	ctx, span := h.tp.Tracer(serverName).Start(c.Request().Context(), "Handler.Position")
	defer span.End()

	cmd := new(positionCommand)
	err := bindCommand(c, cmd)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, errBindFailed)
		return err
	}

	span.SetAttributes(
		attribute.String(attrDroneID, cmd.DroneID),
		attribute.Float64(attrPositionAltitude, float64(cmd.Altitude)),
		attribute.Float64(attrPositionSpeed, float64(cmd.CurrentSpeed)),
	)
	evt := &event.PositionChanged{
		DroneID:         cmd.DroneID,
		Longitude:       cmd.Longitude,
		Latitude:        cmd.Latitude,
		Altitude:        cmd.Altitude,
		CurrentSpeed:    cmd.CurrentSpeed,
		HeadingCardinal: cmd.HeadingCardinal,
		ReceivedAt:      time.Now(),
	}

	err = h.dispatchCommand(ctx, event.PositionUpdate, evt)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, evt)
}
