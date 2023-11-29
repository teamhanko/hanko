package login

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/flow_api/flow/passcode"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type ContinueToPasscodeConfirmationForRecovery struct {
	shared.Action
}

func (a ContinueToPasscodeConfirmationForRecovery) GetName() flowpilot.ActionName {
	return ActionContinueToPasscodeConfirmationRecovery
}

func (a ContinueToPasscodeConfirmationForRecovery) GetDescription() string {
	return "Send a recovery passcode code via email."
}

func (a ContinueToPasscodeConfirmationForRecovery) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	if !deps.Cfg.Passcode.Enabled || !c.Stash().Get("email").Exists() {
		c.SuspendAction()
	}
}

func (a ContinueToPasscodeConfirmationForRecovery) Execute(c flowpilot.ExecutionContext) error {
	if err := c.Stash().Set("passcode_template", "recovery"); err != nil {
		return fmt.Errorf("failed to set passcode_template to stash: %w", err)
	}

	return c.StartSubFlow(passcode.StateConfirmation, StateNewPasswordPrompt)
}
