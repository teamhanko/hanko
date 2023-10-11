package actions

import (
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

func NewSubmitPasscode() SubmitPasscode {
	return SubmitPasscode{}
}

type SubmitPasscode struct{}

func (m SubmitPasscode) GetName() flowpilot.ActionName {
	return common.ActionSubmitPasscode
}

func (m SubmitPasscode) GetDescription() string {
	return "Enter a passcode."
}

func (m SubmitPasscode) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(flowpilot.EmailInput("code").Required(true).Preserve(false))
	// TODO:
}

func (m SubmitPasscode) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	// TODO:

	return c.ContinueFlow(common.StateSuccess)
}
