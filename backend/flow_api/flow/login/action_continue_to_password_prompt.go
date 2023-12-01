package login

import (
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type ContinueToPasswordPrompt struct{}

func (a ContinueToPasswordPrompt) GetName() flowpilot.ActionName {
	return ActionContinueToPasswordPrompt
}

func (a ContinueToPasswordPrompt) GetDescription() string {
	return "Continue to the password login."
}

func (a ContinueToPasswordPrompt) Initialize(_ flowpilot.InitializationContext) {}

func (a ContinueToPasswordPrompt) Execute(c flowpilot.ExecutionContext) error {
	return c.ContinueFlow(StatePasswordPrompt)
}
