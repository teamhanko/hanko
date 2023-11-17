package login

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api/passcode"
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type ContinueToPasscodeConfirmationRecovery struct {
	cfg config.Config
}

func (a ContinueToPasscodeConfirmationRecovery) GetName() flowpilot.ActionName {
	return shared.ActionContinueToPasscodeConfirmationRecovery
}

func (a ContinueToPasscodeConfirmationRecovery) GetDescription() string {
	return "Send a recovery passcode code via email."
}

func (a ContinueToPasscodeConfirmationRecovery) Initialize(c flowpilot.InitializationContext) {
	if !a.cfg.Passcode.Enabled || !c.Stash().Get("email").Exists() {
		c.SuspendAction()
	}
}

func (a ContinueToPasscodeConfirmationRecovery) Execute(c flowpilot.ExecutionContext) error {
	if err := c.Stash().Set("passcode_template", "recovery"); err != nil {
		return fmt.Errorf("failed to set passcode_template to stash: %w", err)
	}

	return c.StartSubFlow(passcode.StatePasscodeConfirmation, StateLoginPasswordRecovery)
}
