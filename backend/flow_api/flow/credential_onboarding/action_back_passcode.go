package credential_onboarding

import (
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type BackPasskey struct {
	shared.Action
}

func (a BackPasskey) GetName() flowpilot.ActionName {
	return shared.ActionBack
}

func (a BackPasskey) GetDescription() string {
	return "Navigate one step back."
}

func (a BackPasskey) Initialize(c flowpilot.InitializationContext) {
	if previousState, _ := c.GetPreviousState(); previousState != nil {
		passwordOnboardingDone := *previousState == shared.StatePasswordCreation &&
			c.Stash().Get("user_has_password").Bool()

		userDetailsOnboardingDone := (*previousState == shared.StateOnboardingUsername && c.Stash().Get("username").Exists()) ||
			(*previousState == shared.StateOnboardingEmail && c.Stash().Get("email").Exists())

		if *previousState == shared.StatePasscodeConfirmation || passwordOnboardingDone || userDetailsOnboardingDone {
			c.SuspendAction()
		}
	}

}

func (a BackPasskey) Execute(c flowpilot.ExecutionContext) error {
	return c.ContinueToPreviousState()
}

func (a BackPasskey) Finalize(c flowpilot.FinalizationContext) error {
	return nil
}
