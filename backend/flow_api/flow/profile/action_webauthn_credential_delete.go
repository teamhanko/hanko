package profile

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type WebauthnCredentialDelete struct {
	shared.Action
}

func (a WebauthnCredentialDelete) GetName() flowpilot.ActionName {
	return ActionWebauthnCredentialDelete
}

func (a WebauthnCredentialDelete) GetDescription() string {
	return "Delete a Webauthn credential."
}

func (a WebauthnCredentialDelete) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(flowpilot.StringInput("passkey_id").Required(true).Hidden(true))
}

func (a WebauthnCredentialDelete) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	if !c.Stash().Get("user_id").Exists() {
		return c.ContinueFlowWithError(
			c.GetErrorState(),
			flowpilot.ErrorOperationNotPermitted.
				Wrap(errors.New("user_id does not exist")))
	}

	userId := uuid.FromStringOrNil(c.Stash().Get("user_id").String())

	userModel, err := deps.Persister.GetUserPersisterWithConnection(deps.Tx).Get(userId)
	if err != nil {
		return fmt.Errorf("could not fetch user %w", err)
	}

	if userModel == nil {
		return errors.New("user not found")
	}

	webauthnCredentialModel := userModel.GetWebauthnCredentialById(c.Input().Get("passkey_id").String())
	if webauthnCredentialModel == nil {
		return c.ContinueFlowWithError(c.GetCurrentState(), shared.ErrorNotFound)
	}

	if (!deps.Cfg.Password.Enabled && !deps.Cfg.Passcode.Enabled) && len(userModel.WebauthnCredentials) == 1 {
		return c.ContinueFlowWithError(
			c.GetCurrentState(),
			flowpilot.ErrorFlowDiscontinuity.
				Wrap(errors.New("cannot delete credential when webauthn is the only auth method enabled")))
	}

	err = deps.Persister.GetWebauthnCredentialPersisterWithConnection(deps.Tx).Delete(*webauthnCredentialModel)
	if err != nil {
		return fmt.Errorf("could not delete passkey: %w", err)
	}

	return c.ContinueFlow(StateProfileInit)
}
