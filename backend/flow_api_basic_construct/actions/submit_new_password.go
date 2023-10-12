package actions

import (
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

func NewSubmitNewPassword() SubmitNewPassword {
	return SubmitNewPassword{}
}

type SubmitNewPassword struct{}

func (m SubmitNewPassword) GetName() flowpilot.ActionName {
	return common.ActionSubmitNewPassword
}

func (m SubmitNewPassword) GetDescription() string {
	return "Submit a new password."
}

func (m SubmitNewPassword) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(flowpilot.PasswordInput("new_password").Required(true))
	// TODO:
}

func (m SubmitNewPassword) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	// TODO:

	return c.ContinueFlow(common.StateSuccess)
}
