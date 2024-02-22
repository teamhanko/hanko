package handler

import (
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/persistence"
	"net/http"
)

type StatusHandler struct {
	persister persistence.Persister
}

func NewStatusHandler(persister persistence.Persister) *StatusHandler {
	return &StatusHandler{
		persister: persister,
	}
}

func (h *StatusHandler) Status(c echo.Context) error {
	// random query to check DB connectivity
	_, err := h.persister.GetJwkPersister().GetAll()
	if err != nil {
		return c.Render(http.StatusInternalServerError, "status", map[string]bool{"dbError": true})
	}

	return c.Render(http.StatusOK, "status", nil)
}
