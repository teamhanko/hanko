package dto

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
)

func ToHttpError(err error) *echo.HTTPError {
	switch e := err.(type) {
	case *echo.HTTPError:
		return &echo.HTTPError{
			Code:     e.Code,
			Message:  fmt.Sprintf("%v", e.Message),
			Internal: e.Internal,
		}
	default:
		return &echo.HTTPError{
			Code:     http.StatusInternalServerError,
			Message:  http.StatusText(http.StatusInternalServerError),
			Internal: err,
		}
	}
}

type HTTPErrorHandlerConfig struct {
	Debug  bool
	Logger echo.Logger
}

func NewHTTPErrorHandler(config HTTPErrorHandlerConfig) func(err error, c echo.Context) {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		herr := ToHttpError(err)

		code := herr.Code
		message := echo.Map{"code": code, "message": herr.Message}
		if config.Debug {
			message = echo.Map{"code": code, "message": herr.Message, "error": err.Error()}
		}

		// Send response
		if c.Request().Method == http.MethodHead { // Issue https://github.com/labstack/echo/issues/608
			err = c.NoContent(code)
		} else {
			err = c.JSON(code, message)
		}
		if err != nil {
			config.Logger.Error(err)
		}
	}
}
