package actions

import (
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

func NewGetWACreationOptions() GetWACreationOptions {
	return GetWACreationOptions{}
}

type GetWACreationOptions struct{}

func (m GetWACreationOptions) GetName() flowpilot.ActionName {
	return common.ActionGetWACreationOptions
}

func (m GetWACreationOptions) GetDescription() string {
	return "Get creation options to create a webauthn credential."
}

func (m GetWACreationOptions) Initialize(c flowpilot.InitializationContext) {
	// TODO:
}

func (m GetWACreationOptions) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	// TODO:

	return c.ContinueFlow(common.StateSuccess)
}
