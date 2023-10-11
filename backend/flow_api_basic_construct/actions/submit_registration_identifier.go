package actions

import (
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence"
)

func NewSubmitRegistrationIdentifier(persister persistence.Persister, httpContext echo.Context) SubmitRegistrationIdentifier {
	return SubmitRegistrationIdentifier{
		persister,
		httpContext,
	}
}

type SubmitRegistrationIdentifier struct {
	persister   persistence.Persister
	httpContext echo.Context
}

func (m SubmitRegistrationIdentifier) GetName() flowpilot.ActionName {
	return common.ActionSubmitRegistrationIdentifier
}

func (m SubmitRegistrationIdentifier) GetDescription() string {
	return "Enter at least one identifier to register."
}

func (m SubmitRegistrationIdentifier) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(flowpilot.EmailInput("email").Required(true).Preserve(true))
	c.AddInputs(flowpilot.StringInput("username").Required(true).Preserve(true))
	// TODO:
}

func (m SubmitRegistrationIdentifier) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	// TODO:

	return c.ContinueFlow(common.StateSuccess)
}
