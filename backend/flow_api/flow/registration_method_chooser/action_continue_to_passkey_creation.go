package registration_method_chooser

import (
	"github.com/teamhanko/hanko/backend/flow_api/flow/passkey_onboarding"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type ContinueToPasskeyCreation struct {
	shared.Action
}

func (a ContinueToPasskeyCreation) GetName() flowpilot.ActionName {
	return ActionContinueToPasskeyRegistration
}

func (a ContinueToPasskeyCreation) GetDescription() string {
	return "Get creation options to create a webauthn credential."
}

func (a ContinueToPasskeyCreation) Initialize(c flowpilot.InitializationContext) {
	if !c.Stash().Get("webauthn_available").Bool() {
		c.SuspendAction()
	}
}

func (a ContinueToPasskeyCreation) Execute(c flowpilot.ExecutionContext) error {
	return c.StartSubFlow(passkey_onboarding.StateOnboardingCreatePasskey)
}

func (a ContinueToPasskeyCreation) Finalize(c flowpilot.FinalizationContext) error {
	return nil
}
