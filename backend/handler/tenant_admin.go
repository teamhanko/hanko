package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgconn"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/teamhanko/hanko/backend/v2/dto"
	"github.com/teamhanko/hanko/backend/v2/dto/admin"
	"github.com/teamhanko/hanko/backend/v2/pagination"
	"github.com/teamhanko/hanko/backend/v2/persistence"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
	"github.com/teamhanko/hanko/backend/v2/webhooks/events"
	"github.com/teamhanko/hanko/backend/v2/webhooks/utils"
)

type TenantHandlerAdmin struct {
	persister persistence.Persister
}

func NewTenantHandlerAdmin(persister persistence.Persister) *TenantHandlerAdmin {
	return &TenantHandlerAdmin{persister: persister}
}

type TenantListRequest struct {
	PerPage int `query:"per_page"`
	Page    int `query:"page"`
}

// List returns a paginated list of tenants
func (h *TenantHandlerAdmin) List(c echo.Context) error {
	var request TenantListRequest
	err := (&echo.DefaultBinder{}).BindQueryParams(c, &request)
	if err != nil {
		return dto.ToHttpError(err)
	}

	if request.Page == 0 {
		request.Page = 1
	}

	if request.PerPage == 0 {
		request.PerPage = 20
	}

	tenants, err := h.persister.GetTenantPersister().List(request.Page, request.PerPage)
	if err != nil {
		return fmt.Errorf("failed to get list of tenants: %w", err)
	}

	tenantCount, err := h.persister.GetTenantPersister().Count()
	if err != nil {
		return fmt.Errorf("failed to get total count of tenants: %w", err)
	}

	u, _ := url.Parse(fmt.Sprintf("%s://%s%s", c.Scheme(), c.Request().Host, c.Request().RequestURI))

	c.Response().Header().Set("Link", pagination.CreateHeader(u, tenantCount, request.Page, request.PerPage))
	c.Response().Header().Set("X-Total-Count", strconv.FormatInt(int64(tenantCount), 10))

	l := make([]admin.Tenant, len(tenants))
	for i := range tenants {
		l[i] = admin.FromTenantModel(tenants[i])
	}

	return c.JSON(http.StatusOK, l)
}

// Get returns a single tenant by ID
func (h *TenantHandlerAdmin) Get(c echo.Context) error {
	tenantId, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse tenantId as uuid").SetInternal(err)
	}

	tenant, err := h.persister.GetTenantPersister().Get(tenantId)
	if err != nil {
		return fmt.Errorf("failed to get tenant: %w", err)
	}

	if tenant == nil {
		return echo.NewHTTPError(http.StatusNotFound, "tenant not found")
	}

	return c.JSON(http.StatusOK, admin.FromTenantModel(*tenant))
}

// Create creates a new tenant
func (h *TenantHandlerAdmin) Create(c echo.Context) error {
	var body admin.CreateTenant
	if err := (&echo.DefaultBinder{}).BindBody(c, &body); err != nil {
		return dto.ToHttpError(err)
	}

	if err := c.Validate(body); err != nil {
		return dto.ToHttpError(err)
	}

	// Generate ID if not provided
	var tenantId uuid.UUID
	if body.ID != nil && !body.ID.IsNil() {
		tenantId = *body.ID
	} else {
		id, err := uuid.NewV4()
		if err != nil {
			return fmt.Errorf("failed to create new tenantId: %w", err)
		}
		tenantId = id
	}

	// Default enabled to true if not specified
	enabled := true
	if body.Enabled != nil {
		enabled = *body.Enabled
	}

	now := time.Now().UTC()
	tenant := models.Tenant{
		ID:        tenantId,
		Name:      body.Name,
		Slug:      body.Slug,
		Config:    body.Config,
		Enabled:   enabled,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := h.persister.Transaction(func(tx *pop.Connection) error {
		err := h.persister.GetTenantPersisterWithConnection(tx).Create(tenant)
		if err != nil {
			var pgErr *pgconn.PgError
			var mysqlErr *mysql.MySQLError
			if errors.As(err, &pgErr) {
				if pgErr.Code == "23505" {
					return echo.NewHTTPError(http.StatusConflict, "tenant with this ID or slug already exists")
				}
			} else if errors.As(err, &mysqlErr) {
				if mysqlErr.Number == 1062 {
					return echo.NewHTTPError(http.StatusConflict, "tenant with this ID or slug already exists")
				}
			}
			return fmt.Errorf("failed to create tenant: %w", err)
		}

		err = utils.TriggerWebhooks(c, tx, events.TenantCreate, admin.FromTenantModel(tenant))
		if err != nil {
			c.Logger().Warn(err)
		}

		return nil
	})

	if httpError, ok := err.(*echo.HTTPError); ok {
		return httpError
	} else if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusCreated, admin.FromTenantModel(tenant))
}

// Update updates an existing tenant
func (h *TenantHandlerAdmin) Update(c echo.Context) error {
	tenantId, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse tenantId as uuid").SetInternal(err)
	}

	var body admin.UpdateTenant
	if err := (&echo.DefaultBinder{}).BindBody(c, &body); err != nil {
		return dto.ToHttpError(err)
	}

	if err := c.Validate(body); err != nil {
		return dto.ToHttpError(err)
	}

	err = h.persister.Transaction(func(tx *pop.Connection) error {
		p := h.persister.GetTenantPersisterWithConnection(tx)
		tenant, err := p.Get(tenantId)
		if err != nil {
			return fmt.Errorf("failed to get tenant: %w", err)
		}

		if tenant == nil {
			return echo.NewHTTPError(http.StatusNotFound, "tenant not found")
		}

		// Update fields if provided
		if body.Name != nil {
			tenant.Name = *body.Name
		}
		if body.Slug != nil {
			tenant.Slug = *body.Slug
		}
		if body.Config != nil {
			tenant.Config = body.Config
		}
		if body.Enabled != nil {
			tenant.Enabled = *body.Enabled
		}
		tenant.UpdatedAt = time.Now().UTC()

		err = p.Update(*tenant)
		if err != nil {
			var pgErr *pgconn.PgError
			var mysqlErr *mysql.MySQLError
			if errors.As(err, &pgErr) {
				if pgErr.Code == "23505" {
					return echo.NewHTTPError(http.StatusConflict, "tenant with this slug already exists")
				}
			} else if errors.As(err, &mysqlErr) {
				if mysqlErr.Number == 1062 {
					return echo.NewHTTPError(http.StatusConflict, "tenant with this slug already exists")
				}
			}
			return fmt.Errorf("failed to update tenant: %w", err)
		}

		err = utils.TriggerWebhooks(c, tx, events.TenantUpdate, admin.FromTenantModel(*tenant))
		if err != nil {
			c.Logger().Warn(err)
		}

		return nil
	})

	if httpError, ok := err.(*echo.HTTPError); ok {
		return httpError
	} else if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	// Fetch updated tenant
	tenant, err := h.persister.GetTenantPersister().Get(tenantId)
	if err != nil {
		return fmt.Errorf("failed to get tenant: %w", err)
	}

	return c.JSON(http.StatusOK, admin.FromTenantModel(*tenant))
}

// Delete deletes a tenant
func (h *TenantHandlerAdmin) Delete(c echo.Context) error {
	tenantId, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse tenantId as uuid").SetInternal(err)
	}

	err = h.persister.Transaction(func(tx *pop.Connection) error {
		p := h.persister.GetTenantPersisterWithConnection(tx)
		tenant, err := p.Get(tenantId)
		if err != nil {
			return fmt.Errorf("failed to get tenant: %w", err)
		}

		if tenant == nil {
			return echo.NewHTTPError(http.StatusNotFound, "tenant not found")
		}

		err = p.Delete(*tenant)
		if err != nil {
			return fmt.Errorf("failed to delete tenant: %w", err)
		}

		err = utils.TriggerWebhooks(c, tx, events.TenantDelete, admin.FromTenantModel(*tenant))
		if err != nil {
			c.Logger().Warn(err)
		}

		return nil
	})

	if httpError, ok := err.(*echo.HTTPError); ok {
		return httpError
	} else if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err)
	}

	return c.NoContent(http.StatusNoContent)
}
