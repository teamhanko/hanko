package middleware

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/context"
	"github.com/teamhanko/hanko/backend/v2/crypto/jwk"
	"github.com/teamhanko/hanko/backend/v2/persistence"
)

func JWKMiddleware(appConfig config.ApplicationConfig, persister persistence.Persister) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			tenant, err := context.GetTenant(ctx)
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "tenant ID required")
			}

			cfg := config.Config{
				ApplicationConfig: appConfig,
				TenantConfig:      tenant.Config,
			}

			jwkManager, err := jwk.NewManager(cfg, persister, appConfig.MultiTenancy.Enabled)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to create JWK manager: %w", err))
			}
			ctx.Set("jwk_manager", jwkManager)

			return next(ctx)
		}
	}
}
