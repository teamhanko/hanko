package credential_usage

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type ContinueToPasscodeConfirmation struct {
	shared.Action
}

func (a ContinueToPasscodeConfirmation) GetName() flowpilot.ActionName {
	return shared.ActionContinueToPasscodeConfirmation
}

func (a ContinueToPasscodeConfirmation) GetDescription() string {
	return "Send a login passcode code via email."
}

func (a ContinueToPasscodeConfirmation) Initialize(c flowpilot.InitializationContext) {}

func (a ContinueToPasscodeConfirmation) Execute(c flowpilot.ExecutionContext) error {
	// Set only for audit logging purposes.
	if err := c.Stash().Set(shared.StashPathLoginMethod, "passcode"); err != nil {
		return fmt.Errorf("failed to set login_method to stash: %w", err)
	}

	return c.Continue(shared.StatePasscodeConfirmation)
}
