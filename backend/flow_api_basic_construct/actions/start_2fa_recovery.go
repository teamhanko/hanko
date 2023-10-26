package actions

import (
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

func NewStart2FARecovery() Start2FARecovery {
	return Start2FARecovery{}
}

type Start2FARecovery struct{}

func (m Start2FARecovery) GetName() flowpilot.ActionName {
	return common.ActionStart2FARecovery
}

func (m Start2FARecovery) GetDescription() string {
	return "Start a second factor recovery."
}

func (m Start2FARecovery) Initialize(c flowpilot.InitializationContext) {
	// TODO:
}

func (m Start2FARecovery) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	// TODO:

	return c.ContinueFlow(common.StateSuccess)
}
