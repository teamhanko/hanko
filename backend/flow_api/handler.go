package flow_api_basic_construct

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api/login"
	"github.com/teamhanko/hanko/backend/flow_api/registration"
	"github.com/teamhanko/hanko/backend/flow_api/shared"
	"github.com/teamhanko/hanko/backend/flow_api/shared/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/session"
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
	return h.executeFlow(c, registration.Flow)
}

func (h *FlowPilotHandler) LoginFlowHandler(c echo.Context) error {
	return h.executeFlow(c, login.Flow)
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

		flow.Set("dependencies", &shared.Dependencies{
			Cfg:             h.cfg,
			Tx:              tx,
			Persister:       h.persister,
			HttpContext:     c,
			SessionManager:  h.sessionManager,
			PasscodeService: h.passcodeService,
		})

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
