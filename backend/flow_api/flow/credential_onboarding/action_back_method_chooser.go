package credential_onboarding

import (
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type BackCredentialOnboardingMethodChooser struct {
	shared.Action
}

func (a BackCredentialOnboardingMethodChooser) GetName() flowpilot.ActionName {
	return shared.ActionBack
}

func (a BackCredentialOnboardingMethodChooser) GetDescription() string {
	return "Navigate one step back."
}

func (a BackCredentialOnboardingMethodChooser) Initialize(c flowpilot.InitializationContext) {
	if c.GetFlowName() == "login" {
		c.SuspendAction()
	} else if c.GetFlowName() == "registration" {
		previousState, _ := c.GetPreviousState()
		if previousState != nil && *previousState == shared.StatePasscodeConfirmation {
			c.SuspendAction()
		}
	}
}

func (a BackCredentialOnboardingMethodChooser) Execute(c flowpilot.ExecutionContext) error {
	return c.ContinueToPreviousState()
}

func (a BackCredentialOnboardingMethodChooser) Finalize(c flowpilot.FinalizationContext) error {
	return nil
}
