package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/crypto/jwk"
	"github.com/teamhanko/hanko/backend/v2/persistence"
	"github.com/teamhanko/hanko/backend/v2/webhooks"
)

func WebhookMiddleware(cfg *config.Config, jwkManager jwk.Generator, persister persistence.Persister) echo.MiddlewareFunc {
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
