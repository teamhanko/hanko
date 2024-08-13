package profile

import (
	"fmt"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/dto/admin"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/webhooks/events"
	"github.com/teamhanko/hanko/backend/webhooks/utils"
)

type AccountDelete struct {
	shared.Action
}

func (a AccountDelete) GetName() flowpilot.ActionName {
	return shared.ActionAccountDelete
}

func (a AccountDelete) GetDescription() string {
	return "Delete an account."
}

func (a AccountDelete) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	if !deps.Cfg.Account.AllowDeletion {
		c.SuspendAction()
	}
}

func (a AccountDelete) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return c.Error(flowpilot.ErrorOperationNotPermitted)
	}

	err := deps.Persister.GetUserPersisterWithConnection(deps.Tx).Delete(*userModel)
	if err != nil {
		return fmt.Errorf("could not delete user: %w", err)
	}

	err = deps.AuditLogger.CreateWithConnection(
		deps.Tx,
		deps.HttpContext,
		models.AuditLogUserDeleted,
		&models.User{ID: userModel.ID},
		nil,
		auditlog.Detail("flow_id", c.GetFlowID()))

	if err != nil {
		return fmt.Errorf("could not create audit log: %w", err)
	}

	cookie, err := deps.SessionManager.DeleteCookie()
	if err != nil {
		return fmt.Errorf("could not delete cookie: %w", err)
	}

	deps.HttpContext.SetCookie(cookie)

	err = utils.TriggerWebhooks(deps.HttpContext, events.UserDelete, admin.FromUserModel(*userModel))
	if err != nil {
		return fmt.Errorf("failed to trrigger webhook: %w", err)
	}

	return c.Continue(shared.StateProfileAccountDeleted)
}
