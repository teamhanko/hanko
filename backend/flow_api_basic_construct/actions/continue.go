package actions

import (
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

func NewContinue() Continue {
	return Continue{}
}

type Continue struct{}

func (m Continue) GetName() flowpilot.ActionName {
	return common.ActionContinue
}

func (m Continue) GetDescription() string {
	return "Continue flow."
}

func (m Continue) Initialize(c flowpilot.InitializationContext) {
	// TODO:
}

func (m Continue) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	// TODO:

	return c.ContinueFlow(common.StateSuccess)
}
