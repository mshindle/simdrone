package web

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

type StructValidator struct {
	validator *validator.Validate
}

func (cv *StructValidator) Validate(i any) error {
	if err := cv.validator.Struct(i); err != nil {
		// You can return the raw error or wrap it in an echo.HTTPError
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func NewStructValidator() *StructValidator {
	return &StructValidator{
		validator: validator.New(),
	}
}
