package flow_api

import (
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	zeroLogger "github.com/rs/zerolog/log"
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
	"strconv"
	"time"
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
	return h.executeFlow(c, flow.RegistrationFlow.MustBuild())
}

func (h *FlowPilotHandler) LoginFlowHandler(c echo.Context) error {
	return h.executeFlow(c, flow.LoginFlow.MustBuild())
}

func (h *FlowPilotHandler) ProfileFlowHandler(c echo.Context) error {
	return h.executeFlow(c, flow.ProfileFlow.MustBuild())
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
		}
	}

	log := zeroLogger.Info().
		Str("time_unix", strconv.FormatInt(time.Now().Unix(), 10)).
		Str("id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Str("remote_ip", c.RealIP()).Str("host", c.Request().Host).
		Str("method", c.Request().Method).Str("uri", c.Request().RequestURI).
		Str("user_agent", c.Request().UserAgent()).Int("status", flowResult.GetStatus()).
		Str("referer", c.Request().Referer())
	if flowResult.GetResponse().Error != nil {
		log.Str("error", fmt.Sprintf(flowResult.GetResponse().Error.Code))
		if flowResult.GetResponse().Error.Internal != nil {
			log.Str("error_internal", *flowResult.GetResponse().Error.Internal)
		}
	}
	log.Send()

	return c.JSON(flowResult.GetStatus(), flowResult.GetResponse())
}

func init() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
}
