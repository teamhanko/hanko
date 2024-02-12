package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/config"
	hankoJwk "github.com/teamhanko/hanko/backend/crypto/jwk"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/webhooks"
)

func WebhookMiddleware(cfg *config.Config, jwkManager hankoJwk.Manager, persister persistence.WebhookPersister) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {

			manager, err := webhooks.NewManager(cfg, persister, jwkManager, ctx.Logger())
			if err != nil {
				return err
			}

			ctx.Set("webhook_manager", manager)

			return next(ctx)
		}
	}
}
