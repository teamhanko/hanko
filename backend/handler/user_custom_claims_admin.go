package handler

import (
	"fmt"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/v3/context"
	"github.com/teamhanko/hanko/backend/v3/dto/admin"
	"github.com/teamhanko/hanko/backend/v3/persistence"
)

type UserCustomClaimsAdminHandler struct {
	persister persistence.Persister
}

func NewUserCustomClaimsAdminHandler(persister persistence.Persister) *UserCustomClaimsAdminHandler {
	return &UserCustomClaimsAdminHandler{persister: persister}
}

func (h *UserCustomClaimsAdminHandler) GetCustomClaims(c echo.Context) error {
	tenant, err := context.GetTenant(c)
	if err != nil {
		return fmt.Errorf("failed to get tenant from context: %w", err)
	}

	userID, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}

	user, err := h.persister.GetUserPersister().GetByPublicID(userID, tenant.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not fetch user").SetInternal(err)
	}
	if user == nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	customClaimsModel, err := h.persister.GetUserCustomClaimsPersister().Get(user.ID, tenant.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not fetch custom claims").SetInternal(err)
	}

	response, err := admin.NewCustomClaims(customClaimsModel)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "could not build custom claims response").SetInternal(err)
	}

	if response == nil {
		return c.NoContent(http.StatusNoContent)
	}
	return c.JSON(http.StatusOK, response)
}
