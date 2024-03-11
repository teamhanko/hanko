package login

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/flow_api/flow/passcode"
	"github.com/teamhanko/hanko/backend/flow_api/flow/passkey_onboarding"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type ContinueToPasscodeConfirmation struct {
	shared.Action
}

func (a ContinueToPasscodeConfirmation) GetName() flowpilot.ActionName {
	return ActionContinueToPasscodeConfirmation
}

func (a ContinueToPasscodeConfirmation) GetDescription() string {
	return "Send a login passcode code via email."
}

func (a ContinueToPasscodeConfirmation) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	if !deps.Cfg.Passcode.Enabled || !c.Stash().Get("email").Exists() {
		c.SuspendAction()
	}
}

func (a ContinueToPasscodeConfirmation) Execute(c flowpilot.ExecutionContext) error {
	deps := a.GetDeps(c)

	if err := c.Stash().Set("passcode_template", "login"); err != nil {
		return fmt.Errorf("failed to set passcode_template to stash: %w", err)
	}

	// Set only for audit logging purposes.
	if err := c.Stash().Set("login_method", "passcode"); err != nil {
		return fmt.Errorf("failed to set login_method to stash: %w", err)
	}

	if deps.Cfg.Passkey.Onboarding.Enabled && c.Stash().Get("webauthn_available").Bool() {
		return c.StartSubFlow(passcode.StatePasscodeConfirmation, passkey_onboarding.StateOnboardingCreatePasskey, shared.StateSuccess)
	}

	return c.StartSubFlow(passcode.StatePasscodeConfirmation, shared.StateSuccess)
}
func (a ContinueToPasscodeConfirmation) Finalize(c flowpilot.FinalizationContext) error {
	return nil
}
