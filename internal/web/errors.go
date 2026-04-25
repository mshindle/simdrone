package web

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v5"
)

func HTTPErrorHandler(c *echo.Context, err error) {
	code := http.StatusInternalServerError
	message := "Internal Server Error"

	var ok bool
	var he *echo.HTTPError
	var sc echo.HTTPStatusCoder
	if he, ok = errors.AsType[*echo.HTTPError](err); ok {
		code = he.Code
		message = fmt.Sprint(he.Message)
		if message == "" {
			message = http.StatusText(he.Code)
		}
		err = he.Unwrap()
	} else if ok = errors.As(err, &sc); ok {
		code = sc.StatusCode()
		message = err.Error()
	}

	// You can add custom logging or notification logic here
	l := LoggerFromEchoContext(c)
	l.Debug().Err(err).Int("code", code).Str("response", message).Msg("caught http error")

	// Send the final response to the client
	if resp, uErr := echo.UnwrapResponse(c.Response()); uErr == nil {
		if resp.Committed {
			return // response has been already sent to the client by handler or some middleware
		}
	}
	m := map[string]any{"code": code, "message": message}
	if err != nil {
		m["error"] = err.Error()
	}
	_ = c.JSON(code, m)
}
