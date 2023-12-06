package profile

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
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

	if !c.Stash().Get("user_id").Exists() {
		return c.ContinueFlowWithError(
			c.GetErrorState(),
			flowpilot.ErrorOperationNotPermitted.
				Wrap(errors.New("user_id does not exist")))
	}

	userId := uuid.FromStringOrNil(c.Stash().Get("user_id").String())

	userModel, err := deps.Persister.GetUserPersisterWithConnection(deps.Tx).Get(userId)
	if err != nil {
		return fmt.Errorf("could not fetch user: %w", err)
	}

	if userModel == nil {
		return errors.New("user not found")
	}

	if !deps.Cfg.Passcode.Enabled && len(userModel.WebauthnCredentials) == 0 {
		return c.ContinueFlowWithError(
			c.GetCurrentState(),
			flowpilot.ErrorFlowDiscontinuity.
				Wrap(errors.New("cannot delete password when recovery not possible and no webauthn credential is available")))
	}

	passwordCredentialModel, err := deps.Persister.GetPasswordCredentialPersisterWithConnection(deps.Tx).GetByUserID(userId)
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
