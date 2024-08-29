package profile

import (
	"fmt"
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

	canDoWebauthn := deps.Cfg.Passkey.Enabled && len(userModel.WebauthnCredentials) > 0

	if !deps.Cfg.Username.Enabled ||
		!deps.Cfg.Username.Optional ||
		userModel.Username == nil ||
		(len(userModel.Emails) == 0 && !canDoWebauthn) {
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

	usernameModel := &models.Username{ID: userModel.Username.ID}
	err := deps.Persister.GetUsernamePersisterWithConnection(deps.Tx).Delete(usernameModel)
	if err != nil {
		return fmt.Errorf("failed to delete username from db: %w", err)
	}
	deletedUsername := userModel.GetUsername()
	userModel.DeleteUsername()

	err = deps.AuditLogger.CreateWithConnection(
		deps.Tx,
		deps.HttpContext,
		models.AuditLogUsernameDeleted,
		&models.User{ID: userModel.ID},
		nil,
		auditlog.Detail("username", *deletedUsername),
		auditlog.Detail("flow_id", c.GetFlowID()))

	if err != nil {
		return fmt.Errorf("could not create audit log: %w", err)
	}

	return c.Continue(shared.StateProfileInit)
}
