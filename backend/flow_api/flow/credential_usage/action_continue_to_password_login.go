package credential_usage

import (
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type ContinueToPasswordLogin struct {
	shared.Action
}

func (a ContinueToPasswordLogin) GetName() flowpilot.ActionName {
	return shared.ActionContinueToPasswordLogin
}

func (a ContinueToPasswordLogin) GetDescription() string {
	return "Continue to the password login."
}

func (a ContinueToPasswordLogin) Initialize(c flowpilot.InitializationContext) {}

func (a ContinueToPasswordLogin) Execute(c flowpilot.ExecutionContext) error {
	return c.Continue(shared.StateLoginPassword)
}
