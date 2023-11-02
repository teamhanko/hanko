package flow_api_basic_construct

import (
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/flows"
	"github.com/teamhanko/hanko/backend/flow_api_basic_construct/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/session"
	"net/http"
)

func NewHandler(cfg config.Config, persister persistence.Persister, passcodeService services.Passcode, sessionManager session.Manager) *FlowPilotHandler {
	return &FlowPilotHandler{
		persister,
		cfg,
		passcodeService,
		sessionManager,
	}
}

type FlowPilotHandler struct {
	persister       persistence.Persister
	cfg             config.Config
	passcodeService services.Passcode
	sessionManager  session.Manager
}

func (h *FlowPilotHandler) RegistrationFlowHandler(c echo.Context) error {
	registrationFlow, err := flows.NewRegistrationFlow(h.cfg, h.persister, h.passcodeService, h.sessionManager, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(fmt.Errorf("failed to create registration flow: %w", err))
	}

	return h.executeFlow(c, registrationFlow)
}

func (h *FlowPilotHandler) LoginFlowHandler(c echo.Context) error {
	loginFlow, err := flows.NewLoginFlow(h.cfg, h.persister, h.passcodeService, h.sessionManager, c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError).SetInternal(fmt.Errorf("failed to create login flow: %w", err))
	}

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
