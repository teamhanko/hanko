package profile

import (
	"errors"
	"fmt"
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
		c.AddInputs(flowpilot.StringInput("username").Required(true))
	}
}

func (a UsernameSet) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return c.ContinueFlowWithError(c.GetErrorState(), flowpilot.ErrorOperationNotPermitted)
	}

	username := c.Input().Get("username").String()

	// check that username only contains allowed characters
	if !utf8.ValidString(username) {
		c.Input().SetError("username", flowpilot.ErrorValueInvalid.Wrap(errors.New("username contains invalid characters")))
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	userModel.Username = username

	duplicateUser, err := deps.Persister.GetUserPersisterWithConnection(deps.Tx).GetByUsername(userModel.Username)
	if err != nil {
		return fmt.Errorf("failed to get user from db: %w", err)
	}

	if duplicateUser != nil && duplicateUser.ID.String() != userModel.ID.String() {
		c.Input().SetError("username", shared.ErrorUsernameAlreadyExists)
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

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
		auditlog.Detail("username", userModel.Username),
		auditlog.Detail("flow_id", c.GetFlowID()))

	if err != nil {
		return fmt.Errorf("could not create audit log: %w", err)
	}

	return c.ContinueFlow(shared.StateProfileInit)
}

func (a UsernameSet) Finalize(c flowpilot.FinalizationContext) error {
	return nil
}
