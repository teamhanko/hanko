package login

import (
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type ContinueToMethodSelection struct {
}

func (a ContinueToMethodSelection) GetName() flowpilot.ActionName {
	return ActionContinueToMethodSelection
}

func (a ContinueToMethodSelection) GetDescription() string {
	return "Navigates to the login method chooser."
}

func (a ContinueToMethodSelection) Initialize(c flowpilot.InitializationContext) {
	if c.Stash().Get("email").Exists() {
		c.SuspendAction()
	}
}

func (a ContinueToMethodSelection) Execute(c flowpilot.ExecutionContext) error {
	return c.ContinueFlow(StateMethodSelection)
}
