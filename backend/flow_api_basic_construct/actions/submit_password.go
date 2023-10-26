package actions

import (
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

func NewSubmitPassword() SubmitPassword {
	return SubmitPassword{}
}

type SubmitPassword struct{}

func (m SubmitPassword) GetName() flowpilot.ActionName {
	return common.ActionSubmitPassword
}

func (m SubmitPassword) GetDescription() string {
	return "Login with a password."
}

func (m SubmitPassword) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(flowpilot.PasswordInput("password").Required(true))
	// TODO:
}

func (m SubmitPassword) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	// TODO:

	return c.ContinueFlow(common.StateSuccess)
}
