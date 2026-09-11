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

type OrganizationHandlerAdmin struct {
	persister persistence.Persister
}

func NewOrganizationHandlerAdmin(persister persistence.Persister) *OrganizationHandlerAdmin {
	return &OrganizationHandlerAdmin{persister: persister}
}

func (h *OrganizationHandlerAdmin) Create(c echo.Context) error {
	tenant, err := context.GetTenant(c)
	if err != nil {
		return fmt.Errorf("failed to get tenant from context: %w", err)
	}

	var body admin.CreateOrganization
	if err := (&echo.DefaultBinder{}).BindBody(c, &body); err != nil {
		return dto.ToHttpError(err)
	}

	if err := c.Validate(body); err != nil {
		return dto.ToHttpError(err)
	}

	id, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("failed to create new organization id: %w", err)
	}

	now := time.Now()
	organization := models.Organization{
		ID:        id,
		TenantID:  tenant.ID,
		Name:      body.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = h.persister.GetOrganizationPersister().Create(organization)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" {
				return echo.NewHTTPError(http.StatusConflict, fmt.Errorf("failed to create organization '%s': %w", body.Name, fmt.Errorf("organization already exists")))
			}
		} else if mysqlErr, ok2 := errors.AsType[*mysql.MySQLError](err); ok2 {
			if mysqlErr.Number == 1062 {
				return echo.NewHTTPError(http.StatusConflict, fmt.Errorf("failed to create organization '%s': %w", body.Name, fmt.Errorf("organization already exists")))
			}
		}
		return fmt.Errorf("failed to create organization: %w", err)
	}

	return c.JSON(http.StatusOK, admin.FromOrganizationModel(organization))
}

type OrganizationListRequest struct {
	PerPage int `query:"per_page"`
	Page    int `query:"page"`
}

func (h *OrganizationHandlerAdmin) List(c echo.Context) error {
	tenant, err := context.GetTenant(c)
	if err != nil {
		return fmt.Errorf("failed to get tenant from context: %w", err)
	}

	var request OrganizationListRequest
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

	organizations, err := h.persister.GetOrganizationPersister().List(request.Page, request.PerPage, tenant.ID)
	if err != nil {
		return fmt.Errorf("failed to get list of organizations: %w", err)
	}

	organizationCount, err := h.persister.GetOrganizationPersister().Count(tenant.ID)
	if err != nil {
		return fmt.Errorf("failed to get total count of organizations: %w", err)
	}

	u, _ := url.Parse(fmt.Sprintf("%s://%s%s", c.Scheme(), c.Request().Host, c.Request().RequestURI))

	c.Response().Header().Set("Link", pagination.CreateHeader(u, organizationCount, request.Page, request.PerPage))
	c.Response().Header().Set("X-Total-Count", strconv.FormatInt(int64(organizationCount), 10))

	l := make([]admin.Organization, len(organizations))
	for i := range organizations {
		l[i] = admin.FromOrganizationModel(organizations[i])
	}

	return c.JSON(http.StatusOK, l)
}

func (h *OrganizationHandlerAdmin) Get(c echo.Context) error {
	tenant, err := context.GetTenant(c)
	if err != nil {
		return fmt.Errorf("failed to get tenant from context: %w", err)
	}

	organizationId, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse organizationId as uuid").SetInternal(err)
	}

	organization, err := h.persister.GetOrganizationPersister().Get(organizationId, tenant.ID)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}

	if organization == nil {
		return echo.NewHTTPError(http.StatusNotFound, "organization not found")
	}

	return c.JSON(http.StatusOK, admin.FromOrganizationModel(*organization))
}

func (h *OrganizationHandlerAdmin) Patch(c echo.Context) error {
	tenant, err := context.GetTenant(c)
	if err != nil {
		return fmt.Errorf("failed to get tenant from context: %w", err)
	}

	organizationId, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse organizationId as uuid").SetInternal(err)
	}

	var body admin.PatchOrganizationRequest
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

	p := h.persister.GetOrganizationPersister()
	organization, err := p.Get(organizationId, tenant.ID)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}
	if organization == nil {
		return echo.NewHTTPError(http.StatusNotFound, "organization not found")
	}

	if body.Name.Present && body.Name.Value != nil {
		organization.Name = *body.Name.Value
		organization.UpdatedAt = time.Now()

		err = p.Update(*organization)
		if err != nil {
			if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
				if pgErr.Code == "23505" {
					return echo.NewHTTPError(http.StatusConflict, fmt.Errorf("failed to update organization '%s': %w", organization.Name, fmt.Errorf("organization already exists")))
				}
			} else if mysqlErr, ok2 := errors.AsType[*mysql.MySQLError](err); ok2 {
				if mysqlErr.Number == 1062 {
					return echo.NewHTTPError(http.StatusConflict, fmt.Errorf("failed to update organization '%s': %w", organization.Name, fmt.Errorf("organization already exists")))
				}
			}
			return fmt.Errorf("failed to update organization: %w", err)
		}
	}

	return c.JSON(http.StatusOK, admin.FromOrganizationModel(*organization))
}

func (h *OrganizationHandlerAdmin) Delete(c echo.Context) error {
	tenant, err := context.GetTenant(c)
	if err != nil {
		return fmt.Errorf("failed to get tenant from context: %w", err)
	}

	organizationId, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse organizationId as uuid").SetInternal(err)
	}

	p := h.persister.GetOrganizationPersister()
	organization, err := p.Get(organizationId, tenant.ID)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}
	if organization == nil {
		return echo.NewHTTPError(http.StatusNotFound, "organization not found")
	}

	err = p.Delete(*organization)
	if err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	return c.NoContent(http.StatusNoContent)
}

// AddMember adds an existing user to an existing organization.
// org_id and user_id are both path segments identifying the resources
// being linked, so an unrecognized id here is a 404, not a 400 - unlike
// a reference embedded in a request body (see RoleBindingHandlerAdmin).
func (h *OrganizationHandlerAdmin) AddMember(c echo.Context) error {
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

	organization, err := h.persister.GetOrganizationPersister().Get(organizationId, tenant.ID)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}
	if organization == nil {
		return echo.NewHTTPError(http.StatusNotFound, "organization not found")
	}

	user, err := h.persister.GetUserPersister().Get(userId, tenant.ID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}

	id, err := uuid.NewV4()
	if err != nil {
		return fmt.Errorf("failed to create new organization membership id: %w", err)
	}

	membership := models.OrganizationMembership{
		ID:             id,
		TenantID:       tenant.ID,
		UserID:         userId,
		OrganizationID: organizationId,
		CreatedAt:      time.Now(),
	}

	err = h.persister.GetOrganizationMembershipPersister().Create(membership)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" {
				return echo.NewHTTPError(http.StatusConflict, "user is already a member of this organization")
			}
		} else if mysqlErr, ok2 := errors.AsType[*mysql.MySQLError](err); ok2 {
			if mysqlErr.Number == 1062 {
				return echo.NewHTTPError(http.StatusConflict, "user is already a member of this organization")
			}
		}
		return fmt.Errorf("failed to add organization member: %w", err)
	}

	return c.JSON(http.StatusOK, admin.FromOrganizationMembershipModel(membership))
}

// RemoveMember removes a user from an organization. This cascades to
// delete that user's role bindings for the organization (enforced by the
// composite foreign key on role_bindings, see the schema migration).
func (h *OrganizationHandlerAdmin) RemoveMember(c echo.Context) error {
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

	organization, err := h.persister.GetOrganizationPersister().Get(organizationId, tenant.ID)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}
	if organization == nil {
		return echo.NewHTTPError(http.StatusNotFound, "organization not found")
	}

	p := h.persister.GetOrganizationMembershipPersister()
	membership, err := p.Get(userId, organizationId, tenant.ID)
	if err != nil {
		return fmt.Errorf("failed to get organization membership: %w", err)
	}
	if membership == nil {
		return echo.NewHTTPError(http.StatusNotFound, "user is not a member of this organization")
	}

	err = p.Delete(*membership)
	if err != nil {
		return fmt.Errorf("failed to remove organization member: %w", err)
	}

	return c.NoContent(http.StatusNoContent)
}

type ListMembersRequest struct {
	PerPage int `query:"per_page"`
	Page    int `query:"page"`
}

func (h *OrganizationHandlerAdmin) ListMembers(c echo.Context) error {
	tenant, err := context.GetTenant(c)
	if err != nil {
		return fmt.Errorf("failed to get tenant from context: %w", err)
	}

	organizationId, err := uuid.FromString(c.Param("org_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse organizationId as uuid").SetInternal(err)
	}

	organization, err := h.persister.GetOrganizationPersister().Get(organizationId, tenant.ID)
	if err != nil {
		return fmt.Errorf("failed to get organization: %w", err)
	}
	if organization == nil {
		return echo.NewHTTPError(http.StatusNotFound, "organization not found")
	}

	var request ListMembersRequest
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

	p := h.persister.GetOrganizationMembershipPersister()
	memberships, err := p.ListByOrganization(organizationId, request.Page, request.PerPage, tenant.ID)
	if err != nil {
		return fmt.Errorf("failed to get list of organization members: %w", err)
	}

	memberCount, err := p.CountByOrganization(organizationId, tenant.ID)
	if err != nil {
		return fmt.Errorf("failed to get total count of organization members: %w", err)
	}

	u, _ := url.Parse(fmt.Sprintf("%s://%s%s", c.Scheme(), c.Request().Host, c.Request().RequestURI))

	c.Response().Header().Set("Link", pagination.CreateHeader(u, memberCount, request.Page, request.PerPage))
	c.Response().Header().Set("X-Total-Count", strconv.FormatInt(int64(memberCount), 10))

	l := make([]admin.OrganizationMember, len(memberships))
	for i := range memberships {
		l[i] = admin.FromOrganizationMembershipModel(memberships[i])
	}

	return c.JSON(http.StatusOK, l)
}
