package handler

import (
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/persistence"
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
	// TODO: What does "Status" check really? Isn't it the status of the entire API? Why query by tenant?
	_, err := h.persister.GetJwkPersister().GetAll(uuid.FromStringOrNil(config.DefaultTenantID))
	if err != nil {
		return c.Render(http.StatusInternalServerError, "status", map[string]bool{"dbError": true})
	}

	return c.Render(http.StatusOK, "status", nil)
}
