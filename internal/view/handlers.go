package view

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/mshindle/simdrone/internal/event"
	"github.com/mshindle/simdrone/internal/repository"
	"github.com/mshindle/simdrone/internal/web"
)

type repoHandler[T event.DroneEvents] func(context.Context, string) (*T, error)

func lastEventHandler[T event.DroneEvents](lh repoHandler[T]) echo.HandlerFunc {
	return func(c *echo.Context) error {
		droneID, err := echo.PathParam[string](c, "droneID")
		if err != nil {
			return err
		}
		evt, err := lh(c.Request().Context(), droneID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return echo.NewHTTPError(http.StatusBadRequest, "no such drone data")
			}
			return echo.NewHTTPError(http.StatusInternalServerError, "unable to process request").Wrap(err)
		}
		return c.JSON(http.StatusOK, evt)
	}
}

func (srv *Server) activeHandler(c *echo.Context) error {
	l := web.LoggerFromEchoContext(c)
	var drones []string
	var dur = 5 * time.Minute

	err := echo.QueryParamsBinder(c).Duration("d", &dur).BindError()
	if err != nil {
		var bErr *echo.BindingError
		errors.As(err, &bErr)
		l.Error().Str("param", bErr.Field).Strs("value", bErr.Values).Msg("invalid query parameters")
		return echo.NewHTTPError(http.StatusBadRequest, "invalid parameter").Wrap(err)
	}

	drones, err = srv.repo.GetActiveDrones(c.Request().Context(), dur)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return echo.NewHTTPError(http.StatusBadRequest, "no active drones found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "unable to process request").Wrap(err)
	}
	slices.Sort(drones)

	return c.JSON(http.StatusOK, drones)
}
