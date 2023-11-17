package login

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api/passcode"
	"github.com/teamhanko/hanko/backend/flow_api/passkey_onboarding"
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type ContinueToPasscodeConfirmation struct {
	cfg config.Config
}

func (a ContinueToPasscodeConfirmation) GetName() flowpilot.ActionName {
	return shared.ActionContinueToPasscodeConfirmation
}

func (a ContinueToPasscodeConfirmation) GetDescription() string {
	return "Send a login passcode code via email."
}

func (a ContinueToPasscodeConfirmation) Initialize(c flowpilot.InitializationContext) {
	if !a.cfg.Passcode.Enabled || !c.Stash().Get("email").Exists() {
		c.SuspendAction()
	}
}

func (a ContinueToPasscodeConfirmation) Execute(c flowpilot.ExecutionContext) error {
	if err := c.Stash().Set("passcode_template", "login"); err != nil {
		return fmt.Errorf("failed to set passcode_template to stash: %w", err)
	}

	if a.cfg.Passkey.Onboarding.Enabled && c.Stash().Get("webauthn_available").Bool() {
		return c.StartSubFlow(passcode.StatePasscodeConfirmation, passkey_onboarding.StateOnboardingCreatePasskey, shared.StateSuccess)
	}

	return c.StartSubFlow(passcode.StatePasscodeConfirmation, shared.StateSuccess)
}
