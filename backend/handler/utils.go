package handler

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func loadDto[I any](ctx echo.Context) (*I, error) {
	var adminDto I
	err := ctx.Bind(&adminDto)
	if err != nil {
		ctx.Logger().Error(err)
		return nil, echo.NewHTTPError(http.StatusBadRequest, err)
	}

	err = ctx.Validate(adminDto)
	if err != nil {
		ctx.Logger().Error(err)
		return nil, echo.NewHTTPError(http.StatusBadRequest, err)
	}

	return &adminDto, nil
}
