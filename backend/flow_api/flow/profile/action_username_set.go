package profile

import (
	"errors"
	"fmt"
	"github.com/gobuffalo/nulls"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"unicode/utf8"
)

type UsernameSet struct {
	shared.Action
}

func (a UsernameSet) GetName() flowpilot.ActionName {
	return shared.ActionUsernameSet
}

func (a UsernameSet) GetDescription() string {
	return "Set the username of a user."
}

func (a UsernameSet) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	if !deps.Cfg.Username.Enabled {
		c.SuspendAction()
	} else {
		c.AddInputs(flowpilot.StringInput("username").Preserve(true).Required(true).TrimSpace(true))
	}
}

func (a UsernameSet) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return c.Error(flowpilot.ErrorOperationNotPermitted)
	}

	username := c.Input().Get("username").String()

	// check that username only contains allowed characters
	if !utf8.ValidString(username) {
		c.Input().SetError("username", flowpilot.ErrorValueInvalid.Wrap(errors.New("username contains invalid characters")))
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	duplicateUser, err := deps.Persister.GetUserPersisterWithConnection(deps.Tx).GetByUsername(username)
	if err != nil {
		return fmt.Errorf("failed to get user from db: %w", err)
	}

	if duplicateUser != nil && duplicateUser.ID.String() != userModel.ID.String() {
		c.Input().SetError("username", shared.ErrorUsernameAlreadyExists)
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	userModel.Username = nulls.NewString(username)

	err = deps.Persister.GetUserPersisterWithConnection(deps.Tx).Update(*userModel)
	if err != nil {
		return fmt.Errorf("could not update user: %w", err)
	}

	err = deps.AuditLogger.CreateWithConnection(
		deps.Tx,
		deps.HttpContext,
		models.AuditLogUsernameChanged,
		&models.User{ID: userModel.ID},
		nil,
		auditlog.Detail("username", userModel.Username.String),
		auditlog.Detail("flow_id", c.GetFlowID()))

	if err != nil {
		return fmt.Errorf("could not create audit log: %w", err)
	}

	return c.Continue(shared.StateProfileInit)
}
