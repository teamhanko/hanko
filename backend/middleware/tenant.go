package middleware

import (
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/context"
	"github.com/teamhanko/hanko/backend/v3/persistence"
)

func TenantMiddlewareMultitenancy(persister persistence.Persister) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			tenantIDStr := ctx.Param("tenant_id")
			if tenantIDStr == "" {
				return echo.NewHTTPError(http.StatusBadRequest, "tenant ID required")
			}

			tenantID, err := uuid.FromString(tenantIDStr)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "invalid tenant ID format").SetInternal(err)
			}

			tenantModel, err := persister.GetTenantPersister().Get(tenantID)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to load tenant").SetInternal(err)
			}
			if tenantModel == nil {
				return echo.NewHTTPError(http.StatusNotFound, "tenant not found")
			}

			requestTenantConfig, err := config.ParseMultitenancyTenantConfig(tenantModel.Config)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to load tenant config").SetInternal(err)
			}

			tenant := context.Tenant{
				ID:     tenantID,
				Config: *requestTenantConfig,
			}

			ctx.Set("tenant", tenant)

			return next(ctx)
		}
	}
}

func TenantMiddlewareSingleTenant(tenantConfig config.TenantConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {

			tenantID, err := uuid.FromString(config.DefaultTenantID)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "cannot determine default tenant ID").SetInternal(err)
			}

			tenant := context.Tenant{
				ID:     tenantID,
				Config: tenantConfig,
			}

			ctx.Set("tenant", tenant)

			return next(ctx)
		}
	}
}
