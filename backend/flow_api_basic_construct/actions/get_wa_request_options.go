package actions

import (
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

func NewGetWARequestOptions() GetWARequestOptions {
	return GetWARequestOptions{}
}

type GetWARequestOptions struct{}

func (m GetWARequestOptions) GetName() flowpilot.ActionName {
	return common.ActionGetWARequestOptions
}

func (m GetWARequestOptions) GetDescription() string {
	return "Get request options to use a webauthn credential."
}

func (m GetWARequestOptions) Initialize(c flowpilot.InitializationContext) {
	// TODO:
}

func (m GetWARequestOptions) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	// TODO:

	return c.ContinueFlow(common.StateSuccess)
}
