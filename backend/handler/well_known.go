package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/config"
	hankoJwk "github.com/teamhanko/hanko/backend/crypto/jwk"
	dto "github.com/teamhanko/hanko/backend/dto"
	"net/http"
)

type WellKnownHandler struct {
	jwkManager hankoJwk.Manager
	config     dto.PublicConfig
}

func NewWellKnownHandler(config config.Config, jwkManager hankoJwk.Manager) (*WellKnownHandler, error) {
	return &WellKnownHandler{
		config:     dto.FromConfig(config),
		jwkManager: jwkManager,
	}, nil
}

func (h *WellKnownHandler) GetPublicKeys(c echo.Context) error {
	keys, err := h.jwkManager.GetPublicKeys()
	if err != nil {
		return err
	}

	c.Response().Header().Add("Cache-Control", "max-age=600")
	return c.JSON(http.StatusOK, keys)
}

func (h *WellKnownHandler) GetConfig(c echo.Context) error {
	return c.JSON(http.StatusOK, h.config)
}
