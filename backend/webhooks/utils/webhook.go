package utils

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/webhooks"
	"github.com/teamhanko/hanko/backend/webhooks/events"
)

func TriggerWebhooks(ctx echo.Context, evt events.Event, data interface{}) error {
	webhookCtx := ctx.Get("webhook_manager")
	if webhookCtx == nil {
		return fmt.Errorf("unable to load webhooks manager from webhook middleware")
	}

	webhookManager := webhookCtx.(webhooks.Manager)
	webhookManager.Trigger(evt, data)

	return nil

}
