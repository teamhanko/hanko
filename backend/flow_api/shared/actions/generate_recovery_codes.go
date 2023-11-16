package actions

import (
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

func NewGenerateRecoveryCodes() GenerateRecoveryCodes {
	return GenerateRecoveryCodes{}
}

type GenerateRecoveryCodes struct{}

func (m GenerateRecoveryCodes) GetName() flowpilot.ActionName {
	return shared.ActionGenerateRecoveryCodes
}

func (m GenerateRecoveryCodes) GetDescription() string {
	return "Generate recovery codes."
}

func (m GenerateRecoveryCodes) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(flowpilot.StringInput("totp_code").Required(true))
	// TODO:
}

func (m GenerateRecoveryCodes) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	// TODO:

	return c.ContinueFlow(shared.StateSuccess)
}
