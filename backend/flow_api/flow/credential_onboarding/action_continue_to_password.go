package credential_onboarding

import (
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type ContinueToPassword struct {
	shared.Action
}

func (a ContinueToPassword) GetName() flowpilot.ActionName {
	return shared.ActionContinueToPasswordRegistration
}

func (a ContinueToPassword) GetDescription() string {
	return "Register a password credential"
}

func (a ContinueToPassword) Initialize(_ flowpilot.InitializationContext) {}

func (a ContinueToPassword) Execute(c flowpilot.ExecutionContext) error {
	return c.ContinueFlow(shared.StatePasswordCreation)
}
