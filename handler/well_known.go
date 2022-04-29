package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/config"
	"github.com/teamhanko/hanko/crypto/jwk"
	"github.com/teamhanko/hanko/dto"
	"net/http"
)

type WellKnownHandler struct {
	jwkManager jwk.Manager
	config     dto.PublicConfig
}

func NewWellKnownHandler(config config.Config, jwkManager jwk.Manager) (*WellKnownHandler, error) {
	return &WellKnownHandler{
		config:     dto.FromConfig(config),
		jwkManager: jwkManager,
	}, nil
}

func (h *WellKnownHandler) GetPublicKeys(c echo.Context) error {
	keys, err := h.jwkManager.GetPublicKeys()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewApiError(http.StatusInternalServerError))
	}
	return c.JSON(http.StatusOK, keys)
}

func (h *WellKnownHandler) GetConfig(c echo.Context) error {
	return c.JSON(http.StatusOK, h.config)
}
