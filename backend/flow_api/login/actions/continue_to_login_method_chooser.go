package actions

import (
	"github.com/teamhanko/hanko/backend/flow_api/login/states"
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type ContinueToLoginMethodChooser struct {
}

func (a ContinueToLoginMethodChooser) GetName() flowpilot.ActionName {
	return shared.ActionContinueToLoginMethodChooser
}

func (a ContinueToLoginMethodChooser) GetDescription() string {
	return "Navigates to the login method chooser."
}

func (a ContinueToLoginMethodChooser) Initialize(c flowpilot.InitializationContext) {
	if c.Stash().Get("email").Exists() {
		c.SuspendAction()
	}
}

func (a ContinueToLoginMethodChooser) Execute(c flowpilot.ExecutionContext) error {
	return c.ContinueFlow(states.StateLoginMethodChooser)
}
