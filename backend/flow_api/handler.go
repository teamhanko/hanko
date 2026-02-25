package flow_api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	zeroLogger "github.com/rs/zerolog/log"
	"github.com/sethvargo/go-limiter"
	auditlog "github.com/teamhanko/hanko/backend/v2/audit_log"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/dto"
	"github.com/teamhanko/hanko/backend/v2/ee/saml"
	"github.com/teamhanko/hanko/backend/v2/flow_api/flow"
	"github.com/teamhanko/hanko/backend/v2/flow_api/flow/shared"
	"github.com/teamhanko/hanko/backend/v2/flow_api/flow_locker"
	"github.com/teamhanko/hanko/backend/v2/flow_api/services"
	"github.com/teamhanko/hanko/backend/v2/flowpilot"
	"github.com/teamhanko/hanko/backend/v2/mapper"
	"github.com/teamhanko/hanko/backend/v2/persistence"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
	"github.com/teamhanko/hanko/backend/v2/session"
)

type FlowPilotHandler struct {
	Persister                   persistence.Persister
	Cfg                         config.Config
	SecurityNotificationService services.SecurityNotification
	PasscodeService             services.Passcode
	PasswordService             services.Password
	WebauthnService             services.WebauthnService
	SamlService                 saml.Service
	SessionManager              session.Manager
	OTPRateLimiter              limiter.Store
	PasscodeRateLimiter         limiter.Store
	PasswordRateLimiter         limiter.Store
	TokenExchangeRateLimiter    limiter.Store
	AuthenticatorMetadata       mapper.AuthenticatorMetadata
	AuditLogger                 auditlog.Logger
	FlowLocker                  flow_locker.FlowLocker
}

func (h *FlowPilotHandler) RegistrationFlowHandler(c echo.Context) error {
	registrationFlow := flow.NewRegistrationFlow(h.Cfg.Debug)

	var executionOpts []flowpilot.FlowExecutionOption

	if h.Cfg.Session.Binding.Enabled {
		s, err := h.getOrCreateAnonymousSession(c)
		if err != nil {
			flowResult := registrationFlow.ResultFromError(err)
			return c.JSON(flowResult.GetStatus(), flowResult.GetResponse())
		}

		executionOpts = append(executionOpts, flowpilot.WithSession(s.ID))
	}

	return h.executeFlow(c, registrationFlow, executionOpts...)
}

func (h *FlowPilotHandler) LoginFlowHandler(c echo.Context) error {
	loginFlow := flow.NewLoginFlow(h.Cfg.Debug)

	var executionOpts []flowpilot.FlowExecutionOption

	if h.Cfg.Session.Binding.Enabled {
		s, err := h.getOrCreateAnonymousSession(c)
		if err != nil {
			flowResult := loginFlow.ResultFromError(err)
			return c.JSON(flowResult.GetStatus(), flowResult.GetResponse())
		}

		executionOpts = append(executionOpts, flowpilot.WithSession(s.ID))
	}

	return h.executeFlow(c, loginFlow, executionOpts...)
}

func (h *FlowPilotHandler) ProfileFlowHandler(c echo.Context) error {
	profileFlow := flow.NewProfileFlow(h.Cfg.Debug)

	s, err := h.validateSession(c)
	if err != nil {
		flowResult := profileFlow.ResultFromError(err)
		return c.JSON(flowResult.GetStatus(), flowResult.GetResponse())
	}

	if s == nil {
		flowResult := profileFlow.ResultFromError(shared.ErrorUnauthorized)
		return c.JSON(flowResult.GetStatus(), flowResult.GetResponse())
	}

	var executionOpts []flowpilot.FlowExecutionOption
	if h.Cfg.Session.Binding.Enabled {
		executionOpts = append(executionOpts, flowpilot.WithSession(s.ID))
	}

	return h.executeFlow(c, profileFlow, executionOpts...)
}

func (h *FlowPilotHandler) TokenExchangeFlowHandler(c echo.Context) error {
	samlIdPInitiatedLoginFlow := flow.NewTokenExchangeFlow(h.Cfg.Debug)
	return h.executeFlow(c, samlIdPInitiatedLoginFlow)
}

func (h *FlowPilotHandler) getOrCreateAnonymousSession(c echo.Context) (*models.Session, error) {
	// We ignore the error, since the only error returned from "Cookie" is an "ErrNoCookie" but not having
	// the cookie is valid.
	anonSessionCookie, _ := c.Cookie(fmt.Sprintf("%s_anonymous", h.Cfg.Session.Cookie.GetName()))

	var sessionModel *models.Session
	var err error

	if anonSessionCookie != nil && anonSessionCookie.Value != "" {
		anonymousSessionId := uuid.FromStringOrNil(anonSessionCookie.Value)

		sessionModel, err = h.Persister.GetSessionPersister().Get(anonymousSessionId)
		if err != nil {
			return nil, fmt.Errorf("could not fetch session model: %w", err)
		}

		if sessionModel != nil && sessionModel.ExpiresAt != nil && time.Now().UTC().After(*sessionModel.ExpiresAt) {
			return nil, flowpilot.ErrorFlowExpired
		}
	}

	action := c.QueryParam("action")
	// If we have an action query parameter, this request attempts to advance an existing flow.
	// We need to check if this flow can be advanced with the given session.
	if action != "" {
		if sessionModel == nil {
			return nil, flowpilot.ErrorOperationNotPermitted.Wrap(
				errors.New("session binding mismatch: flow cannot advance given the current session"),
			)
		}

		flowID, err := extractFlowID(action)
		if err != nil {
			return nil, flowpilot.ErrorFormDataInvalid.Wrap(err)
		}
		if !flowID.IsNil() {
			flowModel, err := h.Persister.GetFlowPersister().GetFlow(flowID)
			if err != nil {
				return nil, fmt.Errorf("could not fetch flow: %w", err)
			}

			if flowModel == nil {
				return nil, shared.ErrorNotFound.Wrap(
					fmt.Errorf("flow with id %s not found", flowID),
				)
			}

			sessionID := *flowModel.SessionID
			if sessionModel.ID.String() != sessionID.String() {
				return nil, flowpilot.ErrorOperationNotPermitted.Wrap(
					errors.New("session binding mismatch: flow cannot advance given the current session"),
				)
			}
		}

		sessionModel.LastUsed = time.Now().UTC()
		err = h.Persister.GetSessionPersister().Update(*sessionModel)
		if err != nil {
			return nil, fmt.Errorf("could not update session model: %w", err)
		}

		return sessionModel, nil
	}

	// If we are here, we are initializing a flow.
	// If we have a session for the given cookie, we simply use the existing session.
	if sessionModel != nil {
		sessionModel.LastUsed = time.Now().UTC()
		err = h.Persister.GetSessionPersister().Update(*sessionModel)
		if err != nil {
			return nil, fmt.Errorf("could not update session model: %w", err)
		}

		return sessionModel, nil
	}

	// No existing session found, create a new one.
	id, _ := uuid.NewV4()
	now := time.Now().UTC()
	expiresAt := now.Add(24 * time.Hour)

	newSession := models.Session{
		ID:        id,
		UserID:    nulls.UUID{},
		UserAgent: nulls.String{},
		IpAddress: nulls.String{},
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: &expiresAt,
		LastUsed:  now,
	}

	if h.Cfg.Session.AcquireIPAddress {
		newSession.IpAddress = nulls.NewString(c.RealIP())
	}

	if h.Cfg.Session.AcquireUserAgent {
		newSession.UserAgent = nulls.NewString(c.Request().UserAgent())
	}

	err = h.Persister.GetSessionPersister().Create(newSession)
	if err != nil {
		return nil, fmt.Errorf("could not create new anonymous session: %w", err)
	}

	cookie := &http.Cookie{
		Name:     fmt.Sprintf("%s_anonymous", h.Cfg.Session.Cookie.GetName()),
		Value:    newSession.ID.String(),
		Domain:   h.Cfg.Session.Cookie.Domain,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(time.Until(expiresAt).Seconds()),
	}

	c.SetCookie(cookie)

	return &newSession, nil
}

func (h *FlowPilotHandler) validateSession(c echo.Context) (*models.Session, error) {
	lookup := fmt.Sprintf("header:Authorization:Bearer,cookie:%s", h.Cfg.Session.Cookie.GetName())
	extractors, err := echojwt.CreateExtractors(lookup)

	if err != nil {
		return nil, flowpilot.ErrorTechnical.Wrap(err)
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

			// check that the session id is stored in the database
			sessionId, ok := token.Get("session_id")
			if !ok {
				lastTokenErr = errors.New("no session id found in token")
				continue
			}
			sessionID, err := uuid.FromString(sessionId.(string))
			if err != nil {
				lastTokenErr = errors.New("session id has wrong format")
				continue
			}

			sessionModel, err := h.Persister.GetSessionPersister().Get(sessionID)
			if err != nil {
				return nil, fmt.Errorf("failed to get session from database: %w", err)
			}
			if sessionModel == nil {
				lastTokenErr = fmt.Errorf("session id not found in database")
				continue
			}

			// Update lastUsed field
			sessionModel.LastUsed = time.Now().UTC()
			err = h.Persister.GetSessionPersister().Update(*sessionModel)
			if err != nil {
				return nil, dto.ToHttpError(err)
			}

			c.Set("session", token)

			return sessionModel, nil
		}
	}

	if lastTokenErr != nil {
		return nil, shared.ErrorUnauthorized.Wrap(lastTokenErr)
	} else if lastExtractorErr != nil {
		return nil, shared.ErrorUnauthorized.Wrap(lastExtractorErr)
	}

	return nil, nil
}

func (h *FlowPilotHandler) executeFlow(c echo.Context, flow flowpilot.Flow, opts ...flowpilot.FlowExecutionOption) error {
	const queryParamKey = "action"

	var err error
	var inputData flowpilot.InputData
	var flowResult flowpilot.FlowResult
	var unlock func(context.Context) error
	var flowID uuid.UUID

	if c.QueryParam(queryParamKey) != "" {
		err = c.Bind(&inputData)
		if err != nil {
			flowResult = flow.ResultFromError(flowpilot.ErrorTechnical.Wrap(err))
			h.logFlowResult(c, flowResult)
			return c.JSON(flowResult.GetStatus(), flowResult.GetResponse())
		}

		flowID, err = extractFlowID(c.QueryParam(queryParamKey))
		if err != nil {
			flowResult = flow.ResultFromError(flowpilot.ErrorTechnical.Wrap(err))
			h.logFlowResult(c, flowResult)
			return c.JSON(flowResult.GetStatus(), flowResult.GetResponse())
		}

		unlock, err = h.FlowLocker.Lock(c.Request().Context(), flowID)
		if err != nil {
			flowResult = flow.ResultFromError(flowpilot.ErrorTechnical.Wrap(fmt.Errorf("could not acquire lock: %w", err)))
			h.logFlowResult(c, flowResult)
			return c.JSON(flowResult.GetStatus(), flowResult.GetResponse())
		}
	}

	txFunc := func(tx *pop.Connection) error {
		deps := &shared.Dependencies{
			Cfg:                         h.Cfg,
			OTPRateLimiter:              h.OTPRateLimiter,
			PasscodeRateLimiter:         h.PasscodeRateLimiter,
			PasswordRateLimiter:         h.PasswordRateLimiter,
			TokenExchangeRateLimiter:    h.TokenExchangeRateLimiter,
			Tx:                          tx,
			Persister:                   h.Persister,
			HttpContext:                 c,
			SessionManager:              h.SessionManager,
			SecurityNotificationService: h.SecurityNotificationService,
			PasscodeService:             h.PasscodeService,
			PasswordService:             h.PasswordService,
			WebauthnService:             h.WebauthnService,
			SamlService:                 h.SamlService,
			AuthenticatorMetadata:       h.AuthenticatorMetadata,
			AuditLogger:                 h.AuditLogger,
		}

		flow.Set("deps", deps)

		executionOptions := []flowpilot.FlowExecutionOption{
			flowpilot.WithQueryParamKey(queryParamKey),
			flowpilot.WithQueryParamValue(c.QueryParam(queryParamKey)),
			flowpilot.WithInputData(inputData),
			flowpilot.UseCompression(!h.Cfg.Debug),
		}

		executionOptions = append(executionOptions, opts...)

		flowResult, err = flow.Execute(persistence.NewFlowPersister(tx),
			executionOptions...)

		return err
	}

	err = h.Persister.Transaction(txFunc)
	if err != nil {
		flowResult = flow.ResultFromError(err)
	}

	if unlock != nil {
		unlockCtx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
		defer cancel()

		if unlockErr := unlock(unlockCtx); unlockErr != nil {
			uErr := fmt.Errorf("failed to release lock: %w", unlockErr)
			if err != nil {
				flowResult = flow.ResultFromError(errors.Join(err, uErr))
			} else {
				flowResult = flow.ResultFromError(flowpilot.ErrorTechnical.Wrap(uErr))
			}
		}
	}

	h.logFlowResult(c, flowResult)

	return c.JSON(flowResult.GetStatus(), flowResult.GetResponse())
}

func (h *FlowPilotHandler) logFlowResult(c echo.Context, flowResult flowpilot.FlowResult) {
	log := zeroLogger.Info().
		Str("time_unix", strconv.FormatInt(time.Now().Unix(), 10)).
		Str("id", c.Response().Header().Get(echo.HeaderXRequestID)).
		Str("remote_ip", c.RealIP()).Str("host", c.Request().Host).
		Str("method", c.Request().Method).Str("uri", c.Request().RequestURI).
		Str("user_agent", c.Request().UserAgent()).Int("status", flowResult.GetStatus()).
		Str("referer", c.Request().Referer())
	if flowResult.GetResponse().Error != nil {
		log.Str("error", fmt.Sprintf("%s", flowResult.GetResponse().Error.Code))
		if flowResult.GetResponse().Error.Internal != nil {
			log.Str("error_internal", *flowResult.GetResponse().Error.Internal)
		}
	}
	log.Send()
}

func init() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
}

// extractFlowID extracts just the flow ID from "action@flowID" format
func extractFlowID(queryParamValue string) (uuid.UUID, error) {
	parts := strings.Split(queryParamValue, "@")
	if len(parts) != 2 {
		return uuid.Nil, fmt.Errorf("invalid flow id format")
	}
	return uuid.FromString(parts[1])
}
