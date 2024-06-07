package credential_onboarding

import (
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type ContinueToPasskey struct {
	shared.Action
}

func (a ContinueToPasskey) GetName() flowpilot.ActionName {
	return shared.ActionContinueToPasskeyRegistration
}

func (a ContinueToPasskey) GetDescription() string {
	return "Register a WebAuthn credential"
}

func (a ContinueToPasskey) Initialize(_ flowpilot.InitializationContext) {}

func (a ContinueToPasskey) Execute(c flowpilot.ExecutionContext) error {
	return c.ContinueFlow(shared.StateOnboardingCreatePasskey)
}
