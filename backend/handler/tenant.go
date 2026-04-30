package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-sql-driver/mysql"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	koanfJson "github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/dto"
	"github.com/teamhanko/hanko/backend/v2/pagination"
	"github.com/teamhanko/hanko/backend/v2/persistence"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
)

type TenantHandler struct {
	persister persistence.Persister
}

func NewTenantHandler(persister persistence.Persister) *TenantHandler {
	return &TenantHandler{persister: persister}
}

type CreateTenantRequest struct {
	ID     *uuid.UUID      `json:"id"`
	Config json.RawMessage `json:"config" validate:"required"`
}

type UpdateTenantRequest struct {
	Config json.RawMessage `json:"config" validate:"required"`
}

type TenantListRequest struct {
	PerPage int `query:"per_page"`
	Page    int `query:"page"`
}

func (h *TenantHandler) Create(c echo.Context) error {
	var body CreateTenantRequest
	if err := (&echo.DefaultBinder{}).BindBody(c, &body); err != nil {
		return dto.ToHttpError(err)
	}

	if err := c.Validate(body); err != nil {
		return dto.ToHttpError(err)
	}

	// Parse and validate the config using koanf
	cfg, err := h.validateTenantConfig(body.Config)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid config: %v", err))
	}

	tenant := models.Tenant{
		Config: cfg,
	}

	// Set ID if provided, otherwise generate a UUID v4
	if body.ID != nil && !body.ID.IsNil() {
		tenant.ID = *body.ID
	} else {
		// TODO: is uui.Must a good idea here?
		tenant.ID = uuid.Must(uuid.NewV4())
	}

	err = h.persister.GetTenantPersister().Create(tenant)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" {
				return echo.NewHTTPError(http.StatusConflict, fmt.Sprintf("failed to create tenant with id '%v': %s", tenant.ID, "tenant already exists"))
			}
		} else if mysqlErr, ok2 := errors.AsType[*mysql.MySQLError](err); ok2 {
			if mysqlErr.Number == 1062 {
				return echo.NewHTTPError(http.StatusConflict, fmt.Sprintf("failed to create tenant with id '%v': %s", tenant.ID, "tenant already exists"))
			}
		}
		return fmt.Errorf("failed to create tenant: %w", err)
	}

	// TODO: Create JWKS for the tenant in a transaction to ensure atomicity

	// Fetch the created tenant to return
	createdTenant, err := h.persister.GetTenantPersister().Get(tenant.ID)
	if err != nil {
		return fmt.Errorf("failed to get created tenant: %w", err)
	}

	if createdTenant == nil {
		return echo.NewHTTPError(http.StatusNotFound, "tenant not found after creation")
	}

	return c.JSON(http.StatusCreated, createdTenant)
}

func (h *TenantHandler) Get(c echo.Context) error {
	tenantId, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse tenant id as uuid").SetInternal(err)
	}

	tenant, err := h.persister.GetTenantPersister().Get(tenantId)
	if err != nil {
		return fmt.Errorf("failed to get tenant: %w", err)
	}

	if tenant == nil {
		return echo.NewHTTPError(http.StatusNotFound, "tenant not found")
	}

	return c.JSON(http.StatusOK, tenant)
}

func (h *TenantHandler) List(c echo.Context) error {
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

	return c.JSON(http.StatusOK, tenants)
}

func (h *TenantHandler) Update(c echo.Context) error {
	tenantId, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse tenant id as uuid").SetInternal(err)
	}

	var body UpdateTenantRequest
	if err := (&echo.DefaultBinder{}).BindBody(c, &body); err != nil {
		return dto.ToHttpError(err)
	}

	if err := c.Validate(body); err != nil {
		return dto.ToHttpError(err)
	}

	// Parse and validate the config using koanf
	cfg, err := h.validateTenantConfig(body.Config)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid config: %v", err))
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

		tenant.Config = cfg

		err = p.Update(*tenant)
		if err != nil {
			return fmt.Errorf("failed to update tenant: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Fetch the updated tenant to return
	updatedTenant, err := h.persister.GetTenantPersister().Get(tenantId)
	if err != nil {
		return fmt.Errorf("failed to get updated tenant: %w", err)
	}

	if updatedTenant == nil {
		return echo.NewHTTPError(http.StatusNotFound, "tenant not found")
	}

	return c.JSON(http.StatusOK, updatedTenant)
}

func (h *TenantHandler) Delete(c echo.Context) error {
	tenantId, err := uuid.FromString(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse tenant id as uuid").SetInternal(err)
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

		return nil
	})
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// validateTenantConfig parses and validates the tenant config JSON using koanf
func (h *TenantHandler) validateTenantConfig(configJSON json.RawMessage) (json.RawMessage, error) {
	if len(configJSON) == 0 {
		return nil, fmt.Errorf("config cannot be empty")
	}

	k := koanf.New(".")

	// Load the JSON config into koanf
	if err := k.Load(rawbytes.Provider(configJSON), koanfJson.Parser()); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	// Unmarshal into TenantConfig to validate structure
	var tenantConfig = new(config.DefaultTenantConfig())
	if err := k.Unmarshal("", &tenantConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal into TenantConfig: %w", err)
	}

	b, err := k.Marshal(koanfJson.Parser())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config JSON: %w", err)
	}
	return b, nil
}
