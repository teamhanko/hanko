package mfa_creation

import (
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
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
}

func (a ContinueToOTPSecretCreation) Execute(c flowpilot.ExecutionContext) error {
	return c.Continue(shared.StateMFAOTPSecretCreation)
}
