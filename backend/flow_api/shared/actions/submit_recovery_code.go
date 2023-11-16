package actions

import (
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

func NewSubmitRecoveryCode() SubmitRecoveryCode {
	return SubmitRecoveryCode{}
}

type SubmitRecoveryCode struct{}

func (m SubmitRecoveryCode) GetName() flowpilot.ActionName {
	return shared.ActionSubmitRecoveryCode
}

func (m SubmitRecoveryCode) GetDescription() string {
	return "Submit a recovery code."
}

func (m SubmitRecoveryCode) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(flowpilot.StringInput("recovery_code").Required(true))
	// TODO:
}

func (m SubmitRecoveryCode) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	// TODO:

	return c.ContinueFlow(shared.StateSuccess)
}
