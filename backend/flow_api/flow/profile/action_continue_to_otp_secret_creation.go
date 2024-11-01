package profile

import (
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type ContinueToOTPSecretCreation struct {
	shared.Action
}

func (a ContinueToOTPSecretCreation) GetName() flowpilot.ActionName {
	return shared.ActionContinueToOTPSecretCreation
}

func (a ContinueToOTPSecretCreation) GetDescription() string {
	return "Create an OTP secret"
}

func (a ContinueToOTPSecretCreation) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	if !deps.Cfg.MFA.TOTP.Enabled {
		c.SuspendAction()
	}

	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		c.SuspendAction()
		return
	}

	if userModel.OTPSecret != nil {
		c.SuspendAction()
	}
}

func (a ContinueToOTPSecretCreation) Execute(c flowpilot.ExecutionContext) error {
	userModel, ok := c.Get("session_user").(*models.User)
	if !ok {
		return c.Error(flowpilot.ErrorOperationNotPermitted)
	}

	if userModel.Emails != nil {
		_ = c.Stash().Set(shared.StashPathEmail, userModel.Emails.GetPrimary().Address)
	}
	_ = c.Stash().Set(shared.StashPathUsername, userModel.Username)

	return c.Continue(shared.StateMFAOTPSecretCreation, shared.StateProfileInit)
}
