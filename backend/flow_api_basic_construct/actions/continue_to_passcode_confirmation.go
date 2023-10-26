package actions

import (
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

func NewContinueToPasscodeConfirmation() flowpilot.Action {
	return ContinueToPasscodeConfirmation{}
}

type ContinueToPasscodeConfirmation struct{}

func (a ContinueToPasscodeConfirmation) GetName() flowpilot.ActionName {
	return common.ActionContinueToPasscodeConfirmation
}

func (a ContinueToPasscodeConfirmation) GetDescription() string {
	return "Send a passcode code via email."
}

func (a ContinueToPasscodeConfirmation) Initialize(c flowpilot.InitializationContext) {
	if !c.Stash().Get("email").Exists() {
		c.SuspendAction()
	}
}

func (a ContinueToPasscodeConfirmation) Execute(c flowpilot.ExecutionContext) error {
	return c.ContinueFlow(common.StateLoginPasscodeConfirmation)
}
