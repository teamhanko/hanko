package actions

import (
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

func NewSubmitTOTPCode() SubmitTOTPCode {
	return SubmitTOTPCode{}
}

type SubmitTOTPCode struct{}

func (m SubmitTOTPCode) GetName() flowpilot.ActionName {
	return shared.ActionSubmitTOTPCode
}

func (m SubmitTOTPCode) GetDescription() string {
	return "Submit a TOTP code."
}

func (m SubmitTOTPCode) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(flowpilot.StringInput("totp_code").Required(true))
	// TODO:
}

func (m SubmitTOTPCode) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	// TODO:

	return c.ContinueFlow(shared.StateSuccess)
}
