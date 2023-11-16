package actions

import (
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

func NewLoginWithOauth() LoginWithOauth {
	return LoginWithOauth{}
}

type LoginWithOauth struct{}

func (m LoginWithOauth) GetName() flowpilot.ActionName {
	return shared.ActionLoginWithOauth
}

func (m LoginWithOauth) GetDescription() string {
	return "Login with a oauth provider."
}

func (m LoginWithOauth) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(
		flowpilot.StringInput("provider").Required(true),
		flowpilot.StringInput("redirect_url"),
	)
	// TODO:
}

func (m LoginWithOauth) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	// TODO:

	return c.ContinueFlow(shared.StateSuccess)
}
