package flow_api_basic_construct

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/flows"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

func NewHandler(cfg config.Config, persister persistence.Persister, passcodeService *services.Passcode) *FlowPilotHandler {
	return &FlowPilotHandler{
		persister,
		cfg,
		passcodeService,
	}
}

type FlowPilotHandler struct {
	persister       persistence.Persister
	cfg             config.Config
	passcodeService *services.Passcode
}

func (h *FlowPilotHandler) RegistrationFlowHandler(c echo.Context) error {
	registrationFlow := flows.NewRegistrationFlow(h.cfg, h.persister, h.passcodeService, c)

	return h.executeFlow(c, registrationFlow)
}

func (h *FlowPilotHandler) LoginFlowHandler(c echo.Context) error {
	loginFlow := flows.NewLoginFlow(h.cfg)

	return h.executeFlow(c, loginFlow)
}

func (h *FlowPilotHandler) executeFlow(c echo.Context, flow flowpilot.Flow) error {
	actionParam := c.QueryParam("flowpilot_action")

	var body flowpilot.InputData
	err := c.Bind(&body)
	if err != nil {
		result := flow.ResultFromError(flowpilot.ErrorTechnical.Wrap(err))
		return c.JSON(result.Status(), result.Response())
	}

	err = h.persister.Transaction(func(tx *pop.Connection) error {
		db := models.NewFlowDB(tx)

		result, flowPilotErr := flow.Execute(db, flowpilot.WithActionParam(actionParam), flowpilot.WithInputData(body))
		if flowPilotErr != nil {
			return flowPilotErr
		}

		return c.JSON(result.Status(), result.Response())
	})

	if err != nil {
		c.Logger().Errorf("tx error: %v", err)
		result := flow.ResultFromError(err)

		return c.JSON(result.Status(), result.Response())
	}

	return nil // TODO: maybe return TechnicalError or something else
}
