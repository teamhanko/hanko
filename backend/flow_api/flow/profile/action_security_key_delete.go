package profile

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type SecurityKeyDelete struct {
	shared.Action
}

func (a SecurityKeyDelete) GetName() flowpilot.ActionName {
	return shared.ActionSecurityKeyDelete
}

func (a SecurityKeyDelete) GetDescription() string {
	return "Delete a security key"
}

func (a SecurityKeyDelete) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		c.SuspendAction()
		return
	}

	if !deps.Cfg.MFA.Enabled || !deps.Cfg.MFA.SecurityKeys.Enabled {
		c.SuspendAction()
		return
	}

	if len(userModel.GetSecurityKeys()) <= 0 {
		c.SuspendAction()
		return
	}

	if deps.Cfg.MFA.Enabled && !deps.Cfg.MFA.Optional &&
		userModel.OTPSecret == nil && len(userModel.GetSecurityKeys()) == 1 {
		c.SuspendAction()
		return
	}

	c.AddInputs(flowpilot.StringInput("security_key_id").Required(true).Hidden(true))
}

func (a SecurityKeyDelete) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if valid := c.ValidateInputData(); !valid {
		return c.Error(flowpilot.ErrorFormDataInvalid)
	}

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return c.Error(flowpilot.ErrorOperationNotPermitted)
	}

	webauthnCredentialModel := userModel.GetWebauthnCredentialById(c.Input().Get("security_key_id").String())
	if webauthnCredentialModel == nil {
		return c.Error(shared.ErrorNotFound)
	}

	err := deps.Persister.GetWebauthnCredentialPersisterWithConnection(deps.Tx).Delete(*webauthnCredentialModel)
	if err != nil {
		return fmt.Errorf("could not delete security key: %w", err)
	}

	userModel.DeleteWebauthnCredential(webauthnCredentialModel.ID)

	return c.Continue(shared.StateProfileInit)
}
