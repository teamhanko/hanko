package profile

import (
	"fmt"
	"github.com/gobuffalo/nulls"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type UsernameDelete struct {
	shared.Action
}

func (a UsernameDelete) GetName() flowpilot.ActionName {
	return shared.ActionUsernameDelete
}

func (a UsernameDelete) GetDescription() string {
	return "Delete the username of a user."
}

func (a UsernameDelete) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		c.SuspendAction()
		return
	}

	if !deps.Cfg.Username.Enabled ||
		!deps.Cfg.Username.Optional ||
		!userModel.Username.Valid ||
		len(userModel.Emails) < 1 {
		c.SuspendAction()
	}
}

func (a UsernameDelete) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return c.Error(flowpilot.ErrorOperationNotPermitted)
	}

	username := userModel.Username.String
	userModel.Username = nulls.String{Valid: false}
	err := deps.Persister.GetUserPersisterWithConnection(deps.Tx).Update(*userModel)
	if err != nil {
		return fmt.Errorf("failed to get user from db: %w", err)
	}

	err = deps.AuditLogger.CreateWithConnection(
		deps.Tx,
		deps.HttpContext,
		models.AuditLogUsernameDeleted,
		&models.User{ID: userModel.ID},
		nil,
		auditlog.Detail("username", username),
		auditlog.Detail("flow_id", c.GetFlowID()))

	if err != nil {
		return fmt.Errorf("could not create audit log: %w", err)
	}

	return c.Continue(shared.StateProfileInit)
}
