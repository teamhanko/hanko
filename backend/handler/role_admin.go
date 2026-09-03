package handler

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/v3/context"
	"github.com/teamhanko/hanko/backend/v3/dto"
	"github.com/teamhanko/hanko/backend/v3/dto/admin"
	"github.com/teamhanko/hanko/backend/v3/pagination"
	"github.com/teamhanko/hanko/backend/v3/persistence"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

type RoleHandlerAdmin struct {
	persister persistence.Persister
}

func NewRoleHandlerAdmin(persister persistence.Persister) *RoleHandlerAdmin {
	return &RoleHandlerAdmin{persister: persister}
}

func (h *RoleHandlerAdmin) Create(c echo.Context) error {
	tenant, err := context.GetTenant(c)
	if err != nil {
		return fmt.Errorf("failed to get tenant from context: %w", err)
	}

	var body admin.CreateRole
	if err := (&echo.DefaultBinder{}).BindBody(c, &body); err != nil {
		return dto.ToHttpError(err)
	}

	if err := c.Validate(body); err != nil {
		return dto.ToHttpError(err)
	}

	// A slug that parses as a UUID would be permanently unreachable via
	// RolePersister.GetByIDOrSlug, which always tries UUID-parsing first.
	if _, err := uuid.FromString(body.Slug); err == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "slug must not be a valid uuid")
	}

	id, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("failed to create new role id: %w", err)
	}

	now := time.Now()
	role := models.Role{
		ID:        id,
		TenantID:  tenant.ID,
		Slug:      body.Slug,
		Name:      body.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = h.persister.GetRolePersister().Create(role)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" {
				return echo.NewHTTPError(http.StatusConflict, fmt.Errorf("failed to create role '%s': %w", body.Slug, fmt.Errorf("role already exists")))
			}
		} else if mysqlErr, ok2 := errors.AsType[*mysql.MySQLError](err); ok2 {
			if mysqlErr.Number == 1062 {
				return echo.NewHTTPError(http.StatusConflict, fmt.Errorf("failed to create role '%s': %w", body.Slug, fmt.Errorf("role already exists")))
			}
		}
		return fmt.Errorf("failed to create role: %w", err)
	}

	return c.JSON(http.StatusOK, admin.FromRoleModel(role))
}

type RoleListRequest struct {
	PerPage int `query:"per_page"`
	Page    int `query:"page"`
}

func (h *RoleHandlerAdmin) List(c echo.Context) error {
	tenant, err := context.GetTenant(c)
	if err != nil {
		return fmt.Errorf("failed to get tenant from context: %w", err)
	}

	var request RoleListRequest
	err = (&echo.DefaultBinder{}).BindQueryParams(c, &request)
	if err != nil {
		return dto.ToHttpError(err)
	}

	if request.Page == 0 {
		request.Page = 1
	}

	if request.PerPage == 0 {
		request.PerPage = 20
	}

	roles, err := h.persister.GetRolePersister().List(request.Page, request.PerPage, tenant.ID)
	if err != nil {
		return fmt.Errorf("failed to get list of roles: %w", err)
	}

	roleCount, err := h.persister.GetRolePersister().Count(tenant.ID)
	if err != nil {
		return fmt.Errorf("failed to get total count of roles: %w", err)
	}

	u, _ := url.Parse(fmt.Sprintf("%s://%s%s", c.Scheme(), c.Request().Host, c.Request().RequestURI))

	c.Response().Header().Set("Link", pagination.CreateHeader(u, roleCount, request.Page, request.PerPage))
	c.Response().Header().Set("X-Total-Count", strconv.FormatInt(int64(roleCount), 10))

	l := make([]admin.Role, len(roles))
	for i := range roles {
		l[i] = admin.FromRoleModel(roles[i])
	}

	return c.JSON(http.StatusOK, l)
}

func (h *RoleHandlerAdmin) Get(c echo.Context) error {
	tenant, err := context.GetTenant(c)
	if err != nil {
		return fmt.Errorf("failed to get tenant from context: %w", err)
	}

	roleId, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse roleId as uuid").SetInternal(err)
	}

	role, err := h.persister.GetRolePersister().Get(roleId, tenant.ID)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}

	if role == nil {
		return echo.NewHTTPError(http.StatusNotFound, "role not found")
	}

	return c.JSON(http.StatusOK, admin.FromRoleModel(*role))
}

func (h *RoleHandlerAdmin) Patch(c echo.Context) error {
	tenant, err := context.GetTenant(c)
	if err != nil {
		return fmt.Errorf("failed to get tenant from context: %w", err)
	}

	roleId, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse roleId as uuid").SetInternal(err)
	}

	var body admin.PatchRoleRequest
	if err := (&echo.DefaultBinder{}).BindBody(c, &body); err != nil {
		return dto.ToHttpError(err)
	}

	if body.Name.Present {
		if body.Name.Value == nil || strings.TrimSpace(*body.Name.Value) == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "name must be a non-empty string")
		}
		trimmed := strings.TrimSpace(*body.Name.Value)
		body.Name.Value = &trimmed
	}

	p := h.persister.GetRolePersister()
	role, err := p.Get(roleId, tenant.ID)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}
	if role == nil {
		return echo.NewHTTPError(http.StatusNotFound, "role not found")
	}

	if body.Name.Present && body.Name.Value != nil {
		role.Name = *body.Name.Value
		role.UpdatedAt = time.Now()

		err = p.Update(*role)
		if err != nil {
			return fmt.Errorf("failed to update role: %w", err)
		}
	}

	return c.JSON(http.StatusOK, admin.FromRoleModel(*role))
}

func (h *RoleHandlerAdmin) Delete(c echo.Context) error {
	tenant, err := context.GetTenant(c)
	if err != nil {
		return fmt.Errorf("failed to get tenant from context: %w", err)
	}

	roleId, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse roleId as uuid").SetInternal(err)
	}

	p := h.persister.GetRolePersister()
	role, err := p.Get(roleId, tenant.ID)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}
	if role == nil {
		return echo.NewHTTPError(http.StatusNotFound, "role not found")
	}

	// Role deletion does not cascade, unlike organization/user deletion:
	// checked explicitly here for a clean error message, backstopped by
	// the RESTRICT foreign key on role_bindings.role_id in case of a race.
	bindingCount, err := p.CountBindings(roleId, tenant.ID)
	if err != nil {
		return fmt.Errorf("failed to count role bindings: %w", err)
	}
	if bindingCount > 0 {
		return echo.NewHTTPError(http.StatusConflict, fmt.Sprintf("role still has %d binding(s)", bindingCount))
	}

	err = p.Delete(*role)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23503" {
				return echo.NewHTTPError(http.StatusConflict, "role still has bindings")
			}
		} else if mysqlErr, ok2 := errors.AsType[*mysql.MySQLError](err); ok2 {
			if mysqlErr.Number == 1451 {
				return echo.NewHTTPError(http.StatusConflict, "role still has bindings")
			}
		}
		return fmt.Errorf("failed to delete role: %w", err)
	}

	return c.NoContent(http.StatusNoContent)
}
