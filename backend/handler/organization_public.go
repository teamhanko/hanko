package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/teamhanko/hanko/backend/v3/context"
	"github.com/teamhanko/hanko/backend/v3/dto"
	"github.com/teamhanko/hanko/backend/v3/persistence"
)

type OrganizationPublicHandler struct {
	persister persistence.Persister
}

func NewOrganizationPublicHandler(persister persistence.Persister) *OrganizationPublicHandler {
	return &OrganizationPublicHandler{persister: persister}
}

// CheckRole checks whether the currently logged-in user - identified by
// their own session JWT via middleware.Session, not a {user_id} path
// param - holds at least one of the given roles within an organization.
// An unrecognized organization_id and "not a member" both resolve to
// {"has_role": false} rather than an error, so a relying party can't use
// this endpoint to probe which organizations exist.
func (h *OrganizationPublicHandler) CheckRole(c echo.Context) error {
	tenant, err := context.GetTenant(c)
	if err != nil {
		return fmt.Errorf("failed to get tenant from context: %w", err)
	}

	sessionToken, ok := c.Get("session").(jwt.Token)
	if !ok {
		return errors.New("failed to cast session object")
	}
	userId := uuid.FromStringOrNil(sessionToken.Subject())

	var body dto.RoleCheckRequest
	if err := (&echo.DefaultBinder{}).BindBody(c, &body); err != nil {
		return dto.ToHttpError(err)
	}

	if err := c.Validate(body); err != nil {
		return dto.ToHttpError(err)
	}

	membership, err := h.persister.GetOrganizationMembershipPersister().Get(userId, body.OrganizationID, tenant.ID)
	if err != nil {
		return fmt.Errorf("failed to get organization membership: %w", err)
	}
	if membership == nil {
		return c.JSON(http.StatusOK, dto.RoleCheckResponse{HasRole: false})
	}

	rolePersister := h.persister.GetRolePersister()
	roleBindingPersister := h.persister.GetRoleBindingPersister()
	for _, ref := range body.Roles {
		role, err := rolePersister.GetByIDOrSlug(ref, tenant.ID)
		if err != nil {
			return fmt.Errorf("failed to resolve role: %w", err)
		}
		if role == nil {
			continue
		}

		binding, err := roleBindingPersister.Get(userId, role.ID, body.OrganizationID, tenant.ID)
		if err != nil {
			return fmt.Errorf("failed to get role binding: %w", err)
		}
		if binding != nil {
			return c.JSON(http.StatusOK, dto.RoleCheckResponse{HasRole: true})
		}
	}

	return c.JSON(http.StatusOK, dto.RoleCheckResponse{HasRole: false})
}
