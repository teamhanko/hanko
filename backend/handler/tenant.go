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
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/crypto/jwk/local_db"
	"github.com/teamhanko/hanko/backend/v3/dto"
	"github.com/teamhanko/hanko/backend/v3/pagination"
	"github.com/teamhanko/hanko/backend/v3/persistence"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

type TenantHandler struct {
	persister persistence.Persister
	cfg       *config.Config
}

func NewTenantHandler(cfg *config.Config, persister persistence.Persister) *TenantHandler {
	return &TenantHandler{persister: persister, cfg: cfg}
}

type CreateTenantRequest struct {
	ID     string          `json:"id" validate:"required,uuid"`
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

	tenantID, err := uuid.FromString(body.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid tenant ID: %v", err))
	}

	// Create allows a partial config: anything not given is filled in from config.DefaultTenantConfig().
	tenantConfigJSON, tenantConfig, err := h.validateTenantConfigForCreate(body.Config)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid config: %v", err))
	}

	tenant := models.Tenant{
		ID:     tenantID,
		Config: tenantConfigJSON,
	}

	return h.persister.Transaction(func(tx *pop.Connection) error {
		err = h.persister.GetTenantPersisterWithConnection(tx).Create(tenant)
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

		if tenantConfig.Secrets.KeyManagement.Type == "local" {
			manager, err := local_db.NewDefaultManager(h.cfg.SecretKeys, h.persister.GetJwkPersisterWithConnection(tx))
			if err != nil {
				return dto.ToHttpError(err)
			}

			_, err = manager.GenerateKey(tenant.ID)

			if err != nil {
				return fmt.Errorf("failed to create JWKS for tenant: %w", err)
			}
		}

		// Fetch the created tenant to return
		createdTenant, err := h.persister.GetTenantPersisterWithConnection(tx).Get(tenant.ID)
		if err != nil {
			return fmt.Errorf("failed to get created tenant: %w", err)
		}

		if createdTenant == nil {
			return echo.NewHTTPError(http.StatusNotFound, "tenant not found after creation")
		}

		cert, err := h.persister.GetSamlCertificatePersisterWithConnection(tx).GetFirst(tenantID)
		if err != nil {
			return fmt.Errorf("failed to fetch SAML certificate: %w", err)
		}

		if cert == nil {
			cert, err = models.NewSamlCertificate(tenantConfig.Service.Name)
			if err != nil {
				return fmt.Errorf("unable to create SAML certificate: %w", err)
			}

			cert.TenantID = tenantID

			err = h.persister.GetSamlCertificatePersisterWithConnection(tx).Create(cert)
			if err != nil {
				return fmt.Errorf("unable to persist SAML certificate: %w", err)
			}
		}

		return c.JSON(http.StatusCreated, createdTenant)
	})
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

	err = h.persister.Transaction(func(tx *pop.Connection) error {
		p := h.persister.GetTenantPersisterWithConnection(tx)

		tenant, err := p.Get(tenantId)
		if err != nil {
			return fmt.Errorf("failed to get tenant: %w", err)
		}

		if tenant == nil {
			return echo.NewHTTPError(http.StatusNotFound, "tenant not found")
		}

		// PUT merges the given config onto the tenant's current one: a field the request omits
		// keeps its existing value. To clear a field, send its zero value explicitly (e.g. "",
		// false, [], {}) — omission always means "leave as is", never "reset".
		cfg, _, err := h.validateTenantConfigForUpdate(tenant.Config, body.Config)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid config: %v", err))
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

// validateTenantConfigForCreate parses the tenant config JSON using koanf, filling in anything not
// given from config.DefaultTenantConfig(). Creation allows a partial document because every section
// has a usable default.
func (h *TenantHandler) validateTenantConfigForCreate(configJSON json.RawMessage) (json.RawMessage, *config.TenantConfig, error) {
	if len(configJSON) == 0 {
		return nil, nil, fmt.Errorf("config cannot be empty")
	}

	k := koanf.New(".")

	if err := k.Load(rawbytes.Provider(configJSON), koanfJson.Parser()); err != nil {
		return nil, nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	tenantConfig := new(config.DefaultTenantConfig())
	if err := k.Unmarshal("", tenantConfig); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal into TenantConfig: %w", err)
	}

	b, err := h.finalizeTenantConfig(tenantConfig)
	if err != nil {
		return nil, nil, err
	}

	return b, tenantConfig, nil
}

// validateTenantConfigForUpdate parses the tenant config JSON for a PUT, merging it onto the
// tenant's existing config: a field the given JSON omits keeps its current value rather than
// resetting to anything. Both sources are loaded into the same koanf instance, so later (given)
// values override earlier (existing) ones only where the given JSON actually specifies a key.
func (h *TenantHandler) validateTenantConfigForUpdate(existingConfigJSON, configJSON json.RawMessage) (json.RawMessage, *config.TenantConfig, error) {
	if len(configJSON) == 0 {
		return nil, nil, fmt.Errorf("config cannot be empty")
	}

	k := koanf.New(".")

	if err := k.Load(rawbytes.Provider(existingConfigJSON), koanfJson.Parser()); err != nil {
		return nil, nil, fmt.Errorf("failed to parse existing config JSON: %w", err)
	}

	if err := k.Load(rawbytes.Provider(configJSON), koanfJson.Parser()); err != nil {
		return nil, nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	tenantConfig := new(config.TenantConfig)
	if err := k.Unmarshal("", tenantConfig); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal into TenantConfig: %w", err)
	}

	b, err := h.finalizeTenantConfig(tenantConfig)
	if err != nil {
		return nil, nil, err
	}

	return b, tenantConfig, nil
}

// finalizeTenantConfig runs the post-processing and semantic/cross-field validation shared by
// creation and replacement, then marshals the result.
func (h *TenantHandler) finalizeTenantConfig(tenantConfig *config.TenantConfig) (json.RawMessage, error) {
	if err := tenantConfig.PostProcess(); err != nil {
		return nil, fmt.Errorf("failed to post process tenant settings: %w", err)
	}

	cfg := config.Config{
		ApplicationConfig: h.cfg.ApplicationConfig,
		TenantConfig:      *tenantConfig,
	}

	if err := cfg.ValidateTenantAndCrossConfig(); err != nil {
		return nil, fmt.Errorf("failed to validate tenant settings: %w", err)
	}

	b, err := json.Marshal(tenantConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config JSON: %w", err)
	}

	return b, nil
}
