package actions

import (
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

func NewSwitch() Switch {
	return Switch{}
}

type Switch struct{}

func (m Switch) GetName() flowpilot.ActionName {
	return common.ActionSwitch
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

	return c.ContinueFlow(common.StateSuccess)
}
