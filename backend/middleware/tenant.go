package middleware

import (
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/persistence"
)

// TenantMiddleware is a middleware function that retrieves tenant information using the tenant ID from the route params.
// It loads the tenant configuration and sets it on the context for downstream middlewares and handlers to use.
// TODO: maybe return flowErrors instead
func TenantMiddleware(multiTenancy bool, tenantConfig *config.TenantConfig, persister persistence.Persister) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			if multiTenancy {
				tenantIDStr := ctx.Param("tenant_id")
				if tenantIDStr == "" {
					return echo.NewHTTPError(http.StatusBadRequest, "tenant ID required")
				}

				tenantID, err := uuid.FromString(tenantIDStr)
				if err != nil {
					return echo.NewHTTPError(http.StatusBadRequest, "invalid tenant ID format").SetInternal(err)
				}

				tenant, err := persister.GetTenantPersister().Get(tenantID)
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, "failed to load tenant").SetInternal(err)
				}
				if tenant == nil {
					return echo.NewHTTPError(http.StatusNotFound, "tenant not found")
				}

				tenantConfig = new(config.DefaultTenantConfig())
				k := koanf.New(".")

				if err := k.Load(rawbytes.Provider(tenant.Config), json.Parser()); err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, "failed to parse tenant config").SetInternal(err)
				}

				if err := k.Unmarshal("", tenantConfig); err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, "failed to unmarshal tenant config").SetInternal(err)
				}

				if err := tenantConfig.Webauthn.PostProcess(); err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, "failed to post process tenant settings").SetInternal(err)
				}

				ctx.Set("tenant_config", tenantConfig)
				ctx.Set("tenant_id", tenantID)
			} else {
				// Single-tenant mode: use reserved default tenant ID
				defaultID := uuid.Nil
				ctx.Set("tenant_id", defaultID)
				ctx.Set("tenant_config", tenantConfig)
			}
			return next(ctx)
		}
	}
}
