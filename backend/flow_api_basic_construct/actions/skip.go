package actions

import (
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

func NewSkip() Skip {
	return Skip{}
}

type Skip struct{}

func (m Skip) GetName() flowpilot.ActionName {
	return common.ActionSkip
}

func (m Skip) GetDescription() string {
	return "Skip the current state."
}

func (m Skip) Initialize(c flowpilot.InitializationContext) {
	// TODO:
}

func (m Skip) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	// TODO:

	return c.ContinueFlow(common.StateSuccess)
}
