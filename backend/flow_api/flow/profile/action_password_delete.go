package profile

import (
	"errors"
	"fmt"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type PasswordDelete struct {
	shared.Action
}

func (a PasswordDelete) GetName() flowpilot.ActionName {
	return ActionPasswordDelete
}

func (a PasswordDelete) GetDescription() string {
	return "Delete a password."
}

func (a PasswordDelete) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	if !deps.Cfg.Password.Enabled {
		c.SuspendAction()
	}
}

func (a PasswordDelete) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return c.ContinueFlowWithError(c.GetErrorState(), flowpilot.ErrorOperationNotPermitted)
	}

	if !deps.Cfg.Passcode.Enabled && len(userModel.WebauthnCredentials) == 0 {
		return c.ContinueFlowWithError(
			c.GetCurrentState(),
			flowpilot.ErrorFlowDiscontinuity.
				Wrap(errors.New("cannot delete password when recovery not possible and no webauthn credential is available")))
	}

	passwordCredentialModel, err := deps.Persister.GetPasswordCredentialPersisterWithConnection(deps.Tx).GetByUserID(userModel.ID)
	if err != nil {
		return fmt.Errorf("could not fetch password credential: %w", err)
	}

	if passwordCredentialModel == nil {
		return c.ContinueFlow(StateProfileInit)
	}

	err = deps.Persister.GetPasswordCredentialPersisterWithConnection(deps.Tx).Delete(*passwordCredentialModel)
	if err != nil {
		return fmt.Errorf("could not delete password credential: %w", err)
	}

	return c.ContinueFlow(StateProfileInit)
}
