package profile

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
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
	deps := a.GetDeps(c)

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		c.SuspendAction()
		return
	}

	if len(userModel.WebauthnCredentials) == 1 {
		if deps.Cfg.Passcode.Enabled && !deps.Cfg.Password.Enabled {
			if deps.Cfg.Identifier.Email.Optional && len(userModel.Emails) == 0 {
				c.SuspendAction()
				return
			}
		} else if !deps.Cfg.Passcode.Enabled && deps.Cfg.Password.Enabled {
			if userModel.PasswordCredential == nil {
				c.SuspendAction()
				return
			}
		} else {
			if len(userModel.Emails) == 0 && userModel.PasswordCredential == nil {
				c.SuspendAction()
				return
			}
		}
	}

	if len(userModel.WebauthnCredentials) == 0 {
		c.SuspendAction()
		return
	}

	c.AddInputs(flowpilot.StringInput("passkey_id").Required(true).Hidden(true))
}

func (a WebauthnCredentialDelete) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return c.ContinueFlowWithError(c.GetErrorState(), flowpilot.ErrorOperationNotPermitted)
	}

	webauthnCredentialModel := userModel.GetWebauthnCredentialById(c.Input().Get("passkey_id").String())
	if webauthnCredentialModel == nil {
		return c.ContinueFlowWithError(c.GetCurrentState(), shared.ErrorNotFound)
	}

	err := deps.Persister.GetWebauthnCredentialPersisterWithConnection(deps.Tx).Delete(*webauthnCredentialModel)
	if err != nil {
		return fmt.Errorf("could not delete passkey: %w", err)
	}

	return c.ContinueFlow(StateProfileInit)
}
