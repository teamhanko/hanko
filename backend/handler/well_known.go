package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/config"
	hankoJwk "github.com/teamhanko/hanko/crypto/jwk"
	dto "github.com/teamhanko/hanko/dto"
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
	return c.JSON(http.StatusOK, keys)
}

func (h *WellKnownHandler) GetConfig(c echo.Context) error {
	return c.JSON(http.StatusOK, h.config)
}
