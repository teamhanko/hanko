package profile

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type WebauthnCredentialRename struct {
	shared.Action
}

func (a WebauthnCredentialRename) GetName() flowpilot.ActionName {
	return shared.ActionWebauthnCredentialRename
}

func (a WebauthnCredentialRename) GetDescription() string {
	return "Rename a Webauthn credential."
}

func (a WebauthnCredentialRename) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	if !deps.Cfg.Passkey.Enabled {
		c.SuspendAction()
		return
	}

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		c.SuspendAction()
		return
	}

	if len(userModel.WebauthnCredentials) == 0 {
		c.SuspendAction()
		return
	}

	c.AddInputs(flowpilot.StringInput("passkey_id").Required(true).Hidden(true))
	c.AddInputs(flowpilot.StringInput("passkey_name").Required(true))
}

func (a WebauthnCredentialRename) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return c.Error(flowpilot.ErrorOperationNotPermitted)
	}

	webauthnCredentialModel := userModel.GetWebauthnCredentialById(c.Input().Get("passkey_id").String())
	if webauthnCredentialModel == nil {
		return c.Error(shared.ErrorNotFound)
	}

	webauthnCredentialName := c.Input().Get("passkey_name").String()
	webauthnCredentialModel.Name = &webauthnCredentialName

	err := deps.Persister.GetWebauthnCredentialPersisterWithConnection(deps.Tx).Update(*webauthnCredentialModel)
	if err != nil {
		return fmt.Errorf("could not update credential: %w", err)
	}

	return c.Continue(shared.StateProfileInit)
}
