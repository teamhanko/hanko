package flow_api

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/labstack/echo/v4"
	"github.com/sethvargo/go-limiter"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/ee/saml"
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
	SamlService           saml.Service
	SessionManager        session.Manager
	RateLimiter           limiter.Store
	AuthenticatorMetadata mapper.AuthenticatorMetadata
	AuditLogger           auditlog.Logger
}

func (h *FlowPilotHandler) RegistrationFlowHandler(c echo.Context) error {
	return h.executeFlow(c, flow.RegistrationFlow.Debug(h.Cfg.Debug).MustBuild())
}

func (h *FlowPilotHandler) LoginFlowHandler(c echo.Context) error {
	return h.executeFlow(c, flow.LoginFlow.Debug(h.Cfg.Debug).MustBuild())
}

func (h *FlowPilotHandler) ProfileFlowHandler(c echo.Context) error {
	return h.executeFlow(c, flow.ProfileFlow.Debug(h.Cfg.Debug).MustBuild())
}

func (h *FlowPilotHandler) executeFlow(c echo.Context, flow flowpilot.Flow) error {
	const queryParamKey = "action"

	var err error
	var inputData flowpilot.InputData
	var flowResult flowpilot.FlowResult

	txFunc := func(tx *pop.Connection) error {
		deps := &shared.Dependencies{
			Cfg:                   h.Cfg,
			RateLimiter:           h.RateLimiter,
			Tx:                    tx,
			Persister:             h.Persister,
			HttpContext:           c,
			SessionManager:        h.SessionManager,
			PasscodeService:       h.PasscodeService,
			PasswordService:       h.PasswordService,
			WebauthnService:       h.WebauthnService,
			SamlService:           h.SamlService,
			AuthenticatorMetadata: h.AuthenticatorMetadata,
			AuditLogger:           h.AuditLogger,
		}

		flow.Set("deps", deps)

		flowResult, err = flow.Execute(models.NewFlowDB(tx),
			flowpilot.WithQueryParamKey(queryParamKey),
			flowpilot.WithQueryParamValue(c.QueryParam(queryParamKey)),
			flowpilot.WithInputData(inputData))

		return err
	}

	err = c.Bind(&inputData)
	if err != nil {
		flowResult = flow.ResultFromError(flowpilot.ErrorTechnical.Wrap(err))
	} else {
		err = h.Persister.Transaction(txFunc)
		if err != nil {
			flowResult = flow.ResultFromError(err)
			c.Logger().Errorf("failed to handle the request: %v", err)
		}
	}

	return c.JSON(flowResult.GetStatus(), flowResult.GetResponse())
}
