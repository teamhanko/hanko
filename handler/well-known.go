package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/crypto/jwk"
	"net/http"
)

type WellKnownHandler struct {
	jwkManager jwk.Manager
}

func NewWellKnownHandler(jwkManager jwk.Manager) (*WellKnownHandler, error) {
	return &WellKnownHandler{jwkManager: jwkManager}, nil
}

func (h *WellKnownHandler) GetPublicKeys(c echo.Context) error {
	keys, err := h.jwkManager.GetPublicKeys()
	if err != nil {
		c.Error(err)
	}
	return c.JSON(http.StatusOK, keys)
}
