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

// TenantMiddleware is a middleware function that retrieves tenant information using the tenant ID from the request path.
// Only when multi tenancy is enabled.
// TODO: maybe return flowErrors instead
// func TenantMiddleware(cfg *config.ApplicationConfig, tenantConfig *config.TenantConfig, persister persistence.Persister) echo.MiddlewareFunc {
func TenantMiddleware(multiTenancy bool, tenantConfig *config.TenantConfig, persister persistence.Persister) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {

			if multiTenancy {
				tenantId, err := uuid.FromString(ctx.Param("tenant_id"))
				if err != nil {
					return echo.NewHTTPError(http.StatusBadRequest).SetInternal(err)
				}
				tenant, err := persister.GetTenantPersister().Get(tenantId)
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(err)
				}
				if tenant == nil {
					return echo.NewHTTPError(http.StatusNotFound).SetInternal(err)
				}

				defaultTenantConfig := config.DefaultTenantConfig()
				tenantConfig = &defaultTenantConfig
				k := koanf.New(".")

				if err := k.Load(rawbytes.Provider(tenant.Config), json.Parser()); err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, "failed to parse tenant config").SetInternal(err)
				}

				err = k.Unmarshal("", tenantConfig)

				if err := tenantConfig.Webauthn.PostProcess(); err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, "failed to post process tenant settings").SetInternal(err)
				}

				ctx.Set("tenant_config", tenantConfig)
				ctx.Set("tenant_id", &tenantId)
			} else {
				ctx.Set("tenant_id", nil)
				ctx.Set("tenant_config", tenantConfig)
			}
			return next(ctx)
		}
	}
}
