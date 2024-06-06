package credential_onboarding

import (
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type BackPassword struct {
	shared.Action
}

func (a BackPassword) GetName() flowpilot.ActionName {
	return shared.ActionBack
}

func (a BackPassword) GetDescription() string {
	return "Navigate one step back."
}

func (a BackPassword) Initialize(c flowpilot.InitializationContext) {
	if previousState, _ := c.GetPreviousState(); previousState != nil {
		passkeyOnboardingDone := *previousState == shared.StateOnboardingVerifyPasskeyAttestation &&
			c.Stash().Get("user_has_webauthn_credential").Bool()

		userDetailsOnboardingDone := (*previousState == shared.StateOnboardingUsername && c.Stash().Get("username").Exists()) ||
			(*previousState == shared.StateOnboardingEmail && c.Stash().Get("email").Exists())

		if *previousState == shared.StatePasscodeConfirmation || passkeyOnboardingDone || userDetailsOnboardingDone {
			c.SuspendAction()
		}
	}

}

func (a BackPassword) Execute(c flowpilot.ExecutionContext) error {
	return c.ContinueToPreviousState()
}

func (a BackPassword) Finalize(c flowpilot.FinalizationContext) error {
	return nil
}
