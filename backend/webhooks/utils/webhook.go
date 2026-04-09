package utils

import (
	"fmt"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/v2/dto/admin"
	"github.com/teamhanko/hanko/backend/v2/persistence"
	"github.com/teamhanko/hanko/backend/v2/utils"
	"github.com/teamhanko/hanko/backend/v2/webhooks"
	"github.com/teamhanko/hanko/backend/v2/webhooks/events"
)

func TriggerWebhooks(ctx echo.Context, tx *pop.Connection, evt events.Event, data interface{}) error {
	webhookCtx := ctx.Get("webhook_manager")
	if webhookCtx == nil {
		return fmt.Errorf("unable to load webhooks manager from webhook middleware")
	}

	tenantID, err := utils.TenantIDFromContext(ctx)
	if err != nil {
		return fmt.Errorf("invalid tenant identifier: %w", err)
	}

	webhookManager := webhookCtx.(webhooks.Manager)
	webhookManager.Trigger(tx, evt, data, tenantID)

	return nil
}

func NotifyUserChange(ctx echo.Context, tx *pop.Connection, persister persistence.Persister, event events.Event, userId uuid.UUID) {
	tenantID, err := utils.TenantIDFromContext(ctx)
	if err != nil {
		ctx.Logger().Warn(fmt.Errorf("invalid tenant identifier: %w", err))
	}

	updatedUser, err := persister.GetUserPersisterWithConnection(tx).Get(userId, tenantID)
	if err != nil {
		ctx.Logger().Warn(fmt.Errorf("failed to fetch updated user: %w", err))
		return
	}

	user := admin.FromUserModel(*updatedUser)
	user.SetUserAgent(ctx.Request().UserAgent())
	user.SetIPAddress(ctx.RealIP())

	err = TriggerWebhooks(ctx, tx, event, user)
	if err != nil {
		ctx.Logger().Warn(err)
	}
}
