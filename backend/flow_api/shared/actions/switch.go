package actions

import (
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

func NewSwitch() Switch {
	return Switch{}
}

type Switch struct{}

func (m Switch) GetName() flowpilot.ActionName {
	return shared.ActionSwitch
}

func (m Switch) GetDescription() string {
	return "Switch to a different state."
}

func (m Switch) Initialize(c flowpilot.InitializationContext) {
	// TODO:
}

func (m Switch) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	// TODO:

	return c.ContinueFlow(shared.StateSuccess)
}
