package flow_api

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/labstack/echo/v4"
	"github.com/sethvargo/go-limiter"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/flow_api/flow"
	"github.com/teamhanko/hanko/backend/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/flow_api/services"
	"github.com/teamhanko/hanko/backend/flowpilot"
	"github.com/teamhanko/hanko/backend/mapper"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/session"
)

type FlowPilotHandler struct {
	Persister             persistence.Persister
	Cfg                   config.Config
	PasscodeService       services.Passcode
	PasswordService       services.Password
	WebauthnService       services.WebauthnService
	SessionManager        session.Manager
	RateLimiter           limiter.Store
	AuthenticatorMetadata mapper.AuthenticatorMetadata
	AuditLogger           auditlog.Logger
}

func (h *FlowPilotHandler) RegistrationFlowHandler(c echo.Context) error {
	flow := flow.RegistrationFlow.MustBuild()
	return h.executeFlow(c, flow)
}

func (h *FlowPilotHandler) LoginFlowHandler(c echo.Context) error {
	return h.executeFlow(c, flow.LoginFlow.MustBuild())
}

func (h *FlowPilotHandler) ProfileFlowHandler(c echo.Context) error {
	return h.executeFlow(c, flow.ProfileFlow.MustBuild())
}

func (h *FlowPilotHandler) executeFlow(c echo.Context, flow flowpilot.Flow) error {
	actionParam := c.QueryParam("flowpilot_action")

	var body flowpilot.InputData
	err := c.Bind(&body)
	if err != nil {
		result := flow.ResultFromError(flowpilot.ErrorTechnical.Wrap(err))
		return c.JSON(result.Status(), result.Response())
	}

	err = h.Persister.Transaction(func(tx *pop.Connection) error {
		db := models.NewFlowDB(tx)

		flow.Set("dependencies", &shared.Dependencies{
			Cfg:                   h.Cfg,
			RateLimiter:           h.RateLimiter,
			Tx:                    tx,
			Persister:             h.Persister,
			HttpContext:           c,
			SessionManager:        h.SessionManager,
			PasscodeService:       h.PasscodeService,
			PasswordService:       h.PasswordService,
			WebauthnService:       h.WebauthnService,
			AuthenticatorMetadata: h.AuthenticatorMetadata,
			AuditLogger:           h.AuditLogger,
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
