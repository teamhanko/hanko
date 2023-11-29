package login

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/flow_api/flow/passcode"
	"github.com/teamhanko/hanko/backend/flow_api/flow/passkey_onboarding"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type ContinueToPasscodeConfirmationLogin struct {
	shared.Action
}

func (a ContinueToPasscodeConfirmationLogin) GetName() flowpilot.ActionName {
	return ActionContinueToPasscodeConfirmationLogin
}

func (a ContinueToPasscodeConfirmationLogin) GetDescription() string {
	return "Send a login passcode code via email."
}

func (a ContinueToPasscodeConfirmationLogin) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	if !deps.Cfg.Passcode.Enabled || !c.Stash().Get("email").Exists() {
		c.SuspendAction()
	}
}

func (a ContinueToPasscodeConfirmationLogin) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if err := c.Stash().Set("passcode_template", "login"); err != nil {
		return fmt.Errorf("failed to set passcode_template to stash: %w", err)
	}

	if deps.Cfg.Passkey.Onboarding.Enabled && c.Stash().Get("webauthn_available").Bool() {
		return c.StartSubFlow(passcode.StateConfirmation, passkey_onboarding.StateIntroduction, StateSuccess)
	}

	return c.StartSubFlow(passcode.StateConfirmation, StateSuccess)
}
