package login

import (
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/webhooks/events"
	"github.com/teamhanko/hanko/backend/webhooks/utils"
)

type TriggerLoginWebhook struct {
	shared.Action
}

func (h TriggerLoginWebhook) Execute(c flowpilot.HookExecutionContext) error {
	deps := h.GetDeps(c)
	userID := uuid.FromStringOrNil(c.Stash().Get(shared.StashPathUserID).String())
	utils.NotifyUserChange(deps.HttpContext, deps.Tx, deps.Persister, events.UserLogin, userID)
	return nil
}
