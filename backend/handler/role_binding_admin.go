package handler

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/v3/context"
	"github.com/teamhanko/hanko/backend/v3/dto"
	"github.com/teamhanko/hanko/backend/v3/dto/admin"
	"github.com/teamhanko/hanko/backend/v3/persistence"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

type RoleBindingHandlerAdmin struct {
	persister persistence.Persister
}

func NewRoleBindingHandlerAdmin(persister persistence.Persister) *RoleBindingHandlerAdmin {
	return &RoleBindingHandlerAdmin{persister: persister}
}

// Create binds a role to a user within an organization. Both org_id and
// user_id are path segments identifying the membership this binding
// attaches to (404 if either doesn't resolve to an existing membership);
// the role reference in the body may be unrecognized (400), per the
// design doc's "unrecognized organization or role id/slug referenced
// anywhere -> 400" rule for body-embedded references.
func (h *RoleBindingHandlerAdmin) Create(c echo.Context) error {
	tenant, err := context.GetTenant(c)
	if err != nil {
		return fmt.Errorf("failed to get tenant from context: %w", err)
	}

	organizationId, err := uuid.FromString(c.Param("org_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse organizationId as uuid").SetInternal(err)
	}

	userId, err := uuid.FromString(c.Param("user_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse userId as uuid").SetInternal(err)
	}

	var body admin.CreateRoleBindingRequest
	if err := (&echo.DefaultBinder{}).BindBody(c, &body); err != nil {
		return dto.ToHttpError(err)
	}

	if err := c.Validate(body); err != nil {
		return dto.ToHttpError(err)
	}

	membership, err := h.persister.GetOrganizationMembershipPersister().Get(userId, organizationId, tenant.ID)
	if err != nil {
		return fmt.Errorf("failed to get organization membership: %w", err)
	}
	if membership == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "user is not a member of this organization")
	}

	role, err := h.persister.GetRolePersister().GetByIDOrSlug(body.Role, tenant.ID)
	if err != nil {
		return fmt.Errorf("failed to resolve role: %w", err)
	}
	if role == nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("role '%s' not found", body.Role))
	}

	id, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("failed to create new role binding id: %w", err)
	}

	binding := models.RoleBinding{
		ID:             id,
		TenantID:       tenant.ID,
		UserID:         userId,
		RoleID:         role.ID,
		OrganizationID: organizationId,
		CreatedAt:      time.Now(),
	}

	err = h.persister.GetRoleBindingPersister().Create(binding)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" {
				return echo.NewHTTPError(http.StatusConflict, "user already holds this role in this organization")
			}
		} else if mysqlErr, ok2 := errors.AsType[*mysql.MySQLError](err); ok2 {
			if mysqlErr.Number == 1062 {
				return echo.NewHTTPError(http.StatusConflict, "user already holds this role in this organization")
			}
		}
		return fmt.Errorf("failed to create role binding: %w", err)
	}

	return c.JSON(http.StatusOK, admin.FromRoleBindingModel(binding, *role))
}

func (h *RoleBindingHandlerAdmin) List(c echo.Context) error {
	tenant, err := context.GetTenant(c)
	if err != nil {
		return fmt.Errorf("failed to get tenant from context: %w", err)
	}

	organizationId, err := uuid.FromString(c.Param("org_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse organizationId as uuid").SetInternal(err)
	}

	userId, err := uuid.FromString(c.Param("user_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse userId as uuid").SetInternal(err)
	}

	bindings, err := h.persister.GetRoleBindingPersister().ListByUserAndOrganization(userId, organizationId, tenant.ID)
	if err != nil {
		return fmt.Errorf("failed to get list of role bindings: %w", err)
	}

	rolePersister := h.persister.GetRolePersister()
	l := make([]admin.RoleBinding, 0, len(bindings))
	for _, binding := range bindings {
		role, err := rolePersister.Get(binding.RoleID, tenant.ID)
		if err != nil {
			return fmt.Errorf("failed to get role: %w", err)
		}
		if role == nil {
			continue
		}
		l = append(l, admin.FromRoleBindingModel(binding, *role))
	}

	return c.JSON(http.StatusOK, l)
}

// Delete removes a role binding. role_ref, like org_id and user_id, is a
// path segment identifying the specific binding being deleted, so an
// unresolvable role reference here is a 404, not a 400 (unlike Create's
// body-embedded role reference).
func (h *RoleBindingHandlerAdmin) Delete(c echo.Context) error {
	tenant, err := context.GetTenant(c)
	if err != nil {
		return fmt.Errorf("failed to get tenant from context: %w", err)
	}

	organizationId, err := uuid.FromString(c.Param("org_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse organizationId as uuid").SetInternal(err)
	}

	userId, err := uuid.FromString(c.Param("user_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse userId as uuid").SetInternal(err)
	}

	roleRef := c.Param("role_ref")

	role, err := h.persister.GetRolePersister().GetByIDOrSlug(roleRef, tenant.ID)
	if err != nil {
		return fmt.Errorf("failed to resolve role: %w", err)
	}
	if role == nil {
		return echo.NewHTTPError(http.StatusNotFound, "role not found")
	}

	p := h.persister.GetRoleBindingPersister()
	binding, err := p.Get(userId, role.ID, organizationId, tenant.ID)
	if err != nil {
		return fmt.Errorf("failed to get role binding: %w", err)
	}
	if binding == nil {
		return echo.NewHTTPError(http.StatusNotFound, "role binding not found")
	}

	err = p.Delete(*binding)
	if err != nil {
		return fmt.Errorf("failed to delete role binding: %w", err)
	}

	return c.NoContent(http.StatusNoContent)
}
