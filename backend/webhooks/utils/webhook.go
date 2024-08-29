package utils

import (
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/dto/admin"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/webhooks"
	"github.com/teamhanko/hanko/backend/webhooks/events"
)

func TriggerWebhooks(ctx echo.Context, tx *pop.Connection, evt events.Event, data interface{}) error {
	webhookCtx := ctx.Get("webhook_manager")
	if webhookCtx == nil {
		return fmt.Errorf("unable to load webhooks manager from webhook middleware")
	}

	webhookManager := webhookCtx.(webhooks.Manager)
	webhookManager.Trigger(tx, evt, data)

	return nil
}

func NotifyUserChange(ctx echo.Context, tx *pop.Connection, persister persistence.Persister, event events.Event, userId uuid.UUID) {
	updatedUser, err := persister.GetUserPersisterWithConnection(tx).Get(userId)
	if err != nil {
		ctx.Logger().Warn(fmt.Errorf("failed to fetch updated user: %w", err))
		return
	}

	err = TriggerWebhooks(ctx, tx, event, admin.FromUserModel(*updatedUser))
	if err != nil {
		ctx.Logger().Warn(err)
	}
}
