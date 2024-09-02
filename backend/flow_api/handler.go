package flow_api

import (
	"fmt"
	"github.com/gobuffalo/pop/v6"
	echojwt "github.com/labstack/echo-jwt/v4"
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
	Persister                persistence.Persister
	Cfg                      config.Config
	PasscodeService          services.Passcode
	PasswordService          services.Password
	WebauthnService          services.WebauthnService
	SamlService              saml.Service
	SessionManager           session.Manager
	OTPRateLimiter           limiter.Store
	PasscodeRateLimiter      limiter.Store
	PasswordRateLimiter      limiter.Store
	TokenExchangeRateLimiter limiter.Store
	AuthenticatorMetadata    mapper.AuthenticatorMetadata
	AuditLogger              auditlog.Logger
}

func (h *FlowPilotHandler) RegistrationFlowHandler(c echo.Context) error {
	registrationFlow := flow.NewRegistrationFlow(h.Cfg.Debug)
	return h.executeFlow(c, registrationFlow)
}

func (h *FlowPilotHandler) LoginFlowHandler(c echo.Context) error {
	loginFlow := flow.NewLoginFlow(h.Cfg.Debug)
	return h.executeFlow(c, loginFlow)
}

func (h *FlowPilotHandler) ProfileFlowHandler(c echo.Context) error {
	profileFlow := flow.NewProfileFlow(h.Cfg.Debug)

	if err := h.validateSession(c); err != nil {
		flowResult := profileFlow.ResultFromError(err)
		return c.JSON(flowResult.GetStatus(), flowResult.GetResponse())
	}

	return h.executeFlow(c, profileFlow)
}

func (h *FlowPilotHandler) validateSession(c echo.Context) error {
	lookup := fmt.Sprintf("header:Authorization:Bearer,cookie:%s", h.Cfg.Session.Cookie.GetName())
	extractors, err := echojwt.CreateExtractors(lookup)

	if err != nil {
		return flowpilot.ErrorTechnical.Wrap(err)
	}

	var lastExtractorErr, lastTokenErr error
	for _, extractor := range extractors {
		auths, extractorErr := extractor(c)
		if extractorErr != nil {
			lastExtractorErr = extractorErr
			continue
		}
		for _, auth := range auths {
			token, tokenErr := h.SessionManager.Verify(auth)
			if tokenErr != nil {
				lastTokenErr = tokenErr
				continue
			}

			c.Set("session", token)

			return nil
		}
	}

	if lastTokenErr != nil {
		return shared.ErrorUnauthorized.Wrap(lastTokenErr)
	} else if lastExtractorErr != nil {
		return shared.ErrorUnauthorized.Wrap(lastExtractorErr)
	}

	return nil
}

func (h *FlowPilotHandler) executeFlow(c echo.Context, flow flowpilot.Flow) error {
	const queryParamKey = "action"

	var err error
	var inputData flowpilot.InputData
	var flowResult flowpilot.FlowResult

	txFunc := func(tx *pop.Connection) error {
		deps := &shared.Dependencies{
			Cfg:                      h.Cfg,
			OTPRateLimiter:           h.OTPRateLimiter,
			PasscodeRateLimiter:      h.PasscodeRateLimiter,
			PasswordRateLimiter:      h.PasswordRateLimiter,
			TokenExchangeRateLimiter: h.TokenExchangeRateLimiter,
			Tx:                       tx,
			Persister:                h.Persister,
			HttpContext:              c,
			SessionManager:           h.SessionManager,
			PasscodeService:          h.PasscodeService,
			PasswordService:          h.PasswordService,
			WebauthnService:          h.WebauthnService,
			SamlService:              h.SamlService,
			AuthenticatorMetadata:    h.AuthenticatorMetadata,
			AuditLogger:              h.AuditLogger,
		}

		flow.Set("deps", deps)

		flowResult, err = flow.Execute(models.NewFlowDB(tx),
			flowpilot.WithQueryParamKey(queryParamKey),
			flowpilot.WithQueryParamValue(c.QueryParam(queryParamKey)),
			flowpilot.WithInputData(inputData),
			flowpilot.UseCompression(!h.Cfg.Debug))

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
