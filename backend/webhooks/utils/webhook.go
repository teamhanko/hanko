package utils

import (
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/webhooks"
	"github.com/teamhanko/hanko/backend/webhooks/events"
)

func TriggerWebhooks(ctx echo.Context, evts events.Events, data interface{}) {
	webhookManager := ctx.Get("webhook_manager").(*webhooks.Manager)
	webhookManager.Trigger(evts, data)
}
