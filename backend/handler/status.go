package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/persistence"
	"net/http"
)

type StatusHandler struct {
	persister persistence.Persister
	cfg       *config.Config
}

func NewStatusHandler(cfg *config.Config, persister persistence.Persister) *StatusHandler {
	return &StatusHandler{
		persister: persister,
		cfg:       cfg,
	}
}

func (h *StatusHandler) Status(c echo.Context) error {
	// random query to check DB connectivity
	_, err := h.persister.GetJwkPersister().GetAll()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.Render(http.StatusOK, "status", nil)
}
