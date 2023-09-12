package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/session"
	"net/http"
)

type SessionHandler struct {
	enableHeader bool
	cookieName   string
	manager      session.Manager
}

func NewSessionHandler(cfg *config.Config, manager session.Manager) *SessionHandler {
	return &SessionHandler{
		enableHeader: cfg.Session.EnableAuthTokenHeader,
		cookieName:   cfg.Session.Cookie.Name + "refresh",
		manager:      manager,
	}
}

func (handler *SessionHandler) ExchangeRefreshToken(c echo.Context) error {
	token := ""
	if handler.enableHeader {
		token = c.Request().Header.Get("X-Refresh-Token")
	} else {
		c, _ := c.Cookie(handler.cookieName)
		if c != nil {
			token = c.Value
		}
	}

	if token == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "missing refresh token")
	}

	err := handler.manager.ExchangeRefreshToken(token, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid refresh token")
	}

	return c.NoContent(http.StatusOK)
}
