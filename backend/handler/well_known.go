package handler

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/v2/context"
)

type WellKnownHandler struct{}

func NewWellKnownHandler() (*WellKnownHandler, error) {
	return &WellKnownHandler{}, nil
}

func (h *WellKnownHandler) GetPublicKeys(c echo.Context) error {
	tenant, err := context.GetTenant(c)
	if err != nil {
		return fmt.Errorf("failed to get tenant from context: %w", err)
	}

	jwkManager, err := context.GetJwkManager(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get JWK manager from context")
	}

	keys, err := jwkManager.GetPublicKeys(tenant.ID)
	if err != nil {
		return err
	}

	c.Response().Header().Add("Cache-Control", "max-age=600")
	return c.JSON(http.StatusOK, keys)
}
