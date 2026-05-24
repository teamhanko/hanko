package utils

import (
	"fmt"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/v2/dto/admin"
	"github.com/teamhanko/hanko/backend/v2/persistence"
	models "github.com/teamhanko/hanko/backend/v2/persistence/models"
	"github.com/teamhanko/hanko/backend/v2/webhooks"
	"github.com/teamhanko/hanko/backend/v2/webhooks/events"
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

	user := admin.FromUserModel(*updatedUser)
	user.SetUserAgent(ctx.Request().UserAgent())
	user.SetIPAddress(ctx.RealIP())

	err = TriggerWebhooks(ctx, tx, event, user)
	if err != nil {
		ctx.Logger().Warn(err)
	}
}

type SessionWebhookPayload struct {
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`
	Exp       int64  `json:"exp"`
}

func NotifySessionCreate(ctx echo.Context, tx *pop.Connection, session models.Session) {
	payload := SessionWebhookPayload{
		SessionID: session.ID.String(),
		UserID:    session.UserID.String(),
		Exp:       0,
	}
	if session.ExpiresAt != nil {
		payload.Exp = session.ExpiresAt.Unix()
	}

	err := TriggerWebhooks(ctx, tx, events.SessionCreate, payload)
	if err != nil {
		ctx.Logger().Warn(err)
	}
}

func NotifySessionDelete(ctx echo.Context, tx *pop.Connection, session models.Session) {
	payload := SessionWebhookPayload{
		SessionID: session.ID.String(),
		UserID:    session.UserID.String(),
		Exp:       0,
	}
	if session.ExpiresAt != nil {
		payload.Exp = session.ExpiresAt.Unix()
	}

	err := TriggerWebhooks(ctx, tx, events.SessionDelete, payload)
	if err != nil {
		ctx.Logger().Warn(err)
	}
}
