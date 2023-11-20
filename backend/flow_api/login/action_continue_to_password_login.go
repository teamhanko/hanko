package login

import (
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type ContinueToPasswordLogin struct{}

func (a ContinueToPasswordLogin) GetName() flowpilot.ActionName {
	return ActionContinueToPasswordLogin
}

func (a ContinueToPasswordLogin) GetDescription() string {
	return "Continue to the password login."
}

func (a ContinueToPasswordLogin) Initialize(_ flowpilot.InitializationContext) {}

func (a ContinueToPasswordLogin) Execute(c flowpilot.ExecutionContext) error {
	return c.ContinueFlow(StateLoginPassword)
}
