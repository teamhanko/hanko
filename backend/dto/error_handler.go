package dto

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
)

type HTTPError struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	Internal error  `json:"-"` // Stores the error returned by an external dependency
}

// Error makes it compatible with `error` interface
func (he *HTTPError) Error() string {
	if he.Internal == nil {
		return fmt.Sprintf("%s", he.Message)
	}
	return fmt.Sprintf("%s: %s", he.Message, he.Internal)
}

// SetInternal sets error to HTTPError.Internal
// The error will not be returned in the response but will be logged.
// If Debug == true the error will be returned in the response
func (he *HTTPError) SetInternal(err error) *HTTPError {
	he.Internal = err
	return he
}

// Unwrap satisfies the Go 1.13 error wrapper interface.
func (he *HTTPError) Unwrap() error {
	return he.Internal
}

// NewHTTPError creates a new HTTPError instance.
func NewHTTPError(code int, message ...string) *HTTPError {
	he := &HTTPError{
		Code:    code,
		Message: http.StatusText(code),
	}
	if len(message) > 0 {
		he.Message = message[0]
	}

	return he
}

func ToHttpError(err error) *HTTPError {
	switch e := err.(type) {
	case *HTTPError:
		return e
	case *echo.HTTPError:
		return &HTTPError{
			Code:     e.Code,
			Message:  fmt.Sprintf("%v", e.Message),
			Internal: e.Internal,
		}
	default:
		return &HTTPError{
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
