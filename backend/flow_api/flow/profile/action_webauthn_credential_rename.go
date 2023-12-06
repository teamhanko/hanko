package profile

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type WebauthnCredentialRename struct {
	shared.Action
}

func (a WebauthnCredentialRename) GetName() flowpilot.ActionName {
	return ActionWebauthnCredentialRename
}

func (a WebauthnCredentialRename) GetDescription() string {
	return "Rename a Webauthn credential."
}

func (a WebauthnCredentialRename) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(flowpilot.StringInput("passkey_id").Required(true).Hidden(true))
	c.AddInputs(flowpilot.StringInput("passkey_name").Required(true))
}

func (a WebauthnCredentialRename) Execute(c flowpilot.ExecutionContext) error {
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
		return fmt.Errorf("could not fetch user: %w", err)
	}

	if userModel == nil {
		return errors.New("user not found")
	}

	webauthnCredentialId := c.Input().Get("passkey_id").String()
	webauthnCredentialModel, err := deps.Persister.GetWebauthnCredentialPersisterWithConnection(deps.Tx).Get(webauthnCredentialId)
	if err != nil {
		return fmt.Errorf("could not fetch passkey: %w", err)
	}

	webauthnCredentialModel := userModel.GetWebauthnCredentialById(c.Input().Get("passkey_id").String())
	if webauthnCredentialModel == nil {
		return c.ContinueFlowWithError(c.GetCurrentState(), shared.ErrorNotFound)
	}

	webauthnCredentialName := c.Input().Get("passkey_name").String()
	webauthnCredentialModel.Name = &webauthnCredentialName

	err = deps.Persister.GetWebauthnCredentialPersisterWithConnection(deps.Tx).Update(*webauthnCredentialModel)
	if err != nil {
		return fmt.Errorf("could not update credential: %w", err)
	}

	return c.ContinueFlow(StateProfileInit)
}
