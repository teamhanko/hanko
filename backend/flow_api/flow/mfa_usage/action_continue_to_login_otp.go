package mfa_usage

import (
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type ContinueToLoginOTP struct {
	shared.Action
}

func (a ContinueToLoginOTP) GetName() flowpilot.ActionName {
	return shared.ActionContinueToLoginOTP
}

func (a ContinueToLoginOTP) GetDescription() string {
	return "Continues the flow to the OTP login."
}

func (a ContinueToLoginOTP) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	if !deps.Cfg.MFA.TOTP.Enabled || !c.Stash().Get(shared.StashPathUserHasOTPSecret).Bool() {
		c.SuspendAction()
	}
}

func (a ContinueToLoginOTP) Execute(c flowpilot.ExecutionContext) error {
	return c.Continue(shared.StateLoginOTP)
}
