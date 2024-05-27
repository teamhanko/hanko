package credential_usage

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type ContinueToPasscodeConfirmationRecovery struct {
	shared.Action
}

func (a ContinueToPasscodeConfirmationRecovery) GetName() flowpilot.ActionName {
	return shared.ActionContinueToPasscodeConfirmationRecovery
}

func (a ContinueToPasscodeConfirmationRecovery) GetDescription() string {
	return "Send a recovery passcode code via email."
}

func (a ContinueToPasscodeConfirmationRecovery) Initialize(c flowpilot.InitializationContext) {
	deps := a.GetDeps(c)

	if !deps.Cfg.Password.Recovery || !c.Stash().Get("email").Exists() {
		c.SuspendAction()
	}
}

func (a ContinueToPasscodeConfirmationRecovery) Execute(c flowpilot.ExecutionContext) error {
	if err := c.Stash().Set("passcode_template", "recovery"); err != nil {
		return fmt.Errorf("failed to set passcode_template to stash: %w", err)
	}

	return c.StartSubFlow(shared.StatePasscodeConfirmation, shared.StateLoginPasswordRecovery)
}

func (a ContinueToPasscodeConfirmationRecovery) Finalize(c flowpilot.FinalizationContext) error {
	return nil
}
