package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/teamhanko/hanko/config"
	hankoJwk "github.com/teamhanko/hanko/crypto/jwk"
	"github.com/teamhanko/hanko/dto"
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

type jwkSet struct {
	Keys []jwk.Key `json:"keys"`
}

func (h *WellKnownHandler) GetPublicKeys(c echo.Context) error {
	keys, err := h.jwkManager.GetPublicKeys()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.NewApiError(http.StatusInternalServerError))
	}
	set := jwkSet{Keys: keys}
	return c.JSON(http.StatusOK, set)
}

func (h *WellKnownHandler) GetConfig(c echo.Context) error {
	return c.JSON(http.StatusOK, h.config)
}
