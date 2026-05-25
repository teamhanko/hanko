package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/v2/context"
	"github.com/teamhanko/hanko/backend/v2/persistence"
	"github.com/teamhanko/hanko/backend/v2/webhooks"
)

func WebhookMiddleware(persister persistence.Persister) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			tenant, err := context.GetTenant(ctx)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to get tenant from context")
			}

			jwkManager, err := context.GetJwkManager(ctx)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to get JWK manager from context")
			}

			manager, err := webhooks.NewManager(tenant.Config, persister, jwkManager, ctx.Logger())
			if err != nil {
				return err
			}

			ctx.Set("webhook_manager", manager)

			return next(ctx)
		}
	}
}
