package actions

import (
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/common"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence"
)

func NewSubmitLoginIdentifier(persister persistence.Persister, httpContext echo.Context) SubmitLoginIdentifier {
	return SubmitLoginIdentifier{
		persister,
		httpContext,
	}
}

type SubmitLoginIdentifier struct {
	persister   persistence.Persister
	httpContext echo.Context
}

func (m SubmitLoginIdentifier) GetName() flowpilot.ActionName {
	return common.ActionSubmitLoginIdentifier
}

func (m SubmitLoginIdentifier) GetDescription() string {
	return "Enter an identifier to login."
}

func (m SubmitLoginIdentifier) Initialize(c flowpilot.InitializationContext) {
	c.AddInputs(flowpilot.EmailInput("identifier").Required(true).Preserve(true))
	// TODO:
}

func (m SubmitLoginIdentifier) Execute(c flowpilot.ExecutionContext) error {
	if valid := c.ValidateInputData(); !valid {
		return c.ContinueFlowWithError(c.GetCurrentState(), flowpilot.ErrorFormDataInvalid)
	}

	// TODO:

	return c.ContinueFlow(common.StateSuccess)
}
