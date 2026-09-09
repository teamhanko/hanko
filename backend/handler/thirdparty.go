package handler

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/labstack/echo/v4"
	zeroLogger "github.com/rs/zerolog/log"
	auditlog "github.com/teamhanko/hanko/backend/v3/audit_log"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/context"
	"github.com/teamhanko/hanko/backend/v3/dto"
	"github.com/teamhanko/hanko/backend/v3/dto/admin"
	"github.com/teamhanko/hanko/backend/v3/persistence"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
	"github.com/teamhanko/hanko/backend/v3/thirdparty"
	"github.com/teamhanko/hanko/backend/v3/utils"
	webhookUtils "github.com/teamhanko/hanko/backend/v3/webhooks/utils"
	"golang.org/x/oauth2"
)

type ThirdPartyHandler struct {
	appConfig   config.ApplicationConfig
	auditLogger auditlog.Logger
	persister   persistence.Persister
}

func NewThirdPartyHandler(appConfig config.ApplicationConfig, persister persistence.Persister, auditLogger auditlog.Logger) *ThirdPartyHandler {
	return &ThirdPartyHandler{
		appConfig:   appConfig,
		auditLogger: auditLogger,
		persister:   persister,
	}
}

func (h *ThirdPartyHandler) CallbackPost(c echo.Context) error {
	tenant, err := context.GetTenant(c)
	if err != nil {
		return h.redirectError(c, thirdparty.ErrorServer("failed to get tenant from context"), "/error") // TODO:
	}

	q, err := c.FormParams()
	if err != nil {
		return h.redirectError(c, thirdparty.ErrorServer("could not get form parameters"), tenant.Config.ThirdParty.ErrorRedirectURL)
	}
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/thirdparty/callback?%s", q.Encode()))
}

func (h *ThirdPartyHandler) Callback(c echo.Context) error {
	tenant, err := context.GetTenant(c)
	if err != nil {
		return thirdparty.ErrorServer("failed to get tenant from context").WithCause(err)
	}

	var redirectToURL *url.URL
	var accountLinkingResult *thirdparty.AccountLinkingResult
	err = h.persister.Transaction(func(tx *pop.Connection) error {
		var callback dto.ThirdPartyAuthCallback
		terr := c.Bind(&callback)
		if terr != nil {
			return thirdparty.ErrorServer("could not decode request payload").WithCause(terr)
		}

		terr = c.Validate(callback)
		if terr != nil {
			if eerr, ok := terr.(*echo.HTTPError); ok {
				if message, ok2 := eerr.Message.(string); ok2 {
					return thirdparty.ErrorInvalidRequest(message).WithCause(terr)
				} else {
					return thirdparty.ErrorInvalidRequest(terr.Error()).WithCause(terr)
				}
			} else {
				return thirdparty.ErrorInvalidRequest(terr.Error()).WithCause(terr)
			}
		}

		expectedStateCookie, terr := c.Cookie(utils.HankoThirdpartyStateCookie)
		if terr != nil && !errors.Is(terr, http.ErrNoCookie) {
			return thirdparty.ErrorInvalidRequest("could not read state cookie").WithCause(terr)
		}

		var expectedState string
		if expectedStateCookie != nil {
			expectedState = expectedStateCookie.Value
		}
		var state *thirdparty.State
		state, terr = thirdparty.VerifyState(h.appConfig, callback.State, expectedState)
		if terr != nil {
			return thirdparty.ErrorInvalidRequest(terr.Error()).WithCause(terr)
		}

		redirectToURL, terr = url.Parse(state.RedirectTo)
		if terr != nil {
			return thirdparty.ErrorServer("could not parse redirect url").WithCause(terr)
		}

		if callback.HasError() {
			return thirdparty.NewThirdPartyError(callback.Error, callback.ErrorDescription)
		}

		provider, terr := thirdparty.GetProvider(tenant.Config.ThirdParty, state.Provider)
		if terr != nil {
			return thirdparty.ErrorInvalidRequest(terr.Error()).WithCause(terr)
		}

		if callback.AuthCode == "" {
			return thirdparty.ErrorInvalidRequest("auth code missing from request")
		}

		opts := []oauth2.AuthCodeOption{}
		if state.CodeVerifier != "" && provider.ID() != "linkedin" {
			opts = append(opts, oauth2.VerifierOption(state.CodeVerifier))
		}
		oAuthToken, terr := provider.GetOAuthToken(c.Request().Context(), callback.AuthCode, opts...)
		if terr != nil {
			return thirdparty.ErrorInvalidRequest("could not exchange authorization code for access token").WithCause(terr)
		}

		userData, terr := provider.GetUserData(c.Request().Context(), oAuthToken)
		if terr != nil {
			return thirdparty.ErrorInvalidRequest("could not retrieve user data from provider").WithCause(terr)
		}

		linkingResult, terr := thirdparty.LinkAccount(tx, &tenant.Config, h.persister, userData, provider.ID(), false, nil, state.IsFlow, state.UserID, tenant.ID)
		if terr != nil {
			return terr
		}
		accountLinkingResult = linkingResult

		identityModel, err := h.persister.GetIdentityPersisterWithConnection(tx).Get(userData.Metadata.Subject, provider.ID(), tenant.ID)
		if err != nil {
			return thirdparty.ErrorServer("could not get identity").WithCause(err)
		}

		if identityModel != nil && state.UserID != nil && identityModel.UserID != nil && *identityModel.UserID != *state.UserID {
			return thirdparty.ErrorInvalidRequest("identity already exists for a different user")
		}

		tokenOpts := []func(*models.Token){
			models.TokenForFlowAPI(state.IsFlow),
			models.TokenWithIdentityID(identityModel.ID),
			models.TokenUserCreated(linkingResult.UserCreated),
			models.TokenWithLinkUser(state.UserID != nil),
		}
		if state.CodeVerifier != "" {
			tokenOpts = append(tokenOpts, models.TokenPKCESessionVerifier(state.CodeVerifier))
		}
		token, terr := models.NewToken(
			linkingResult.User.ID,
			tenant.ID,
			tokenOpts...,
		)
		if terr != nil {
			return thirdparty.ErrorServer("could not create token").WithCause(terr)
		}

		terr = h.persister.GetTokenPersisterWithConnection(tx).Create(*token)
		if terr != nil {
			return thirdparty.ErrorServer("could not save token to db").WithCause(terr)
		}

		query := redirectToURL.Query()
		query.Add(utils.HankoTokenQuery, token.Value)
		redirectToURL.RawQuery = query.Encode()

		c.SetCookie(&http.Cookie{
			Name:     utils.HankoThirdpartyStateCookie,
			Value:    "",
			Path:     "/",
			Domain:   tenant.Config.Session.Cookie.Domain,
			MaxAge:   -1,
			Secure:   tenant.Config.Session.Cookie.Secure,
			HttpOnly: tenant.Config.Session.Cookie.HttpOnly,
			SameSite: http.SameSiteLaxMode,
		})

		return nil
	})

	errorRedirect := tenant.Config.ThirdParty.ErrorRedirectURL
	if redirectToURL != nil {
		errorRedirect = redirectToURL.String()
	}

	if err != nil {
		return h.redirectError(c, err, errorRedirect)
	}

	err = h.auditLogger.Create(c, accountLinkingResult.Type, accountLinkingResult.User, nil, tenant.ID)
	if err != nil {
		return h.redirectError(c, thirdparty.ErrorServer("could not create audit log").WithCause(err), errorRedirect)
	}

	if accountLinkingResult.WebhookEvent != nil {
		err = webhookUtils.TriggerWebhooks(c, h.persister.GetConnection(), tenant.ID, *accountLinkingResult.WebhookEvent, admin.FromUserModel(*accountLinkingResult.User))
		if err != nil {
			c.Logger().Warn(err)
		}
	}

	return c.Redirect(http.StatusTemporaryRedirect, redirectToURL.String())
}

func (h *ThirdPartyHandler) redirectError(c echo.Context, error error, to string) error {
	// TODO:
	//redirectTo := h.cfg.ThirdParty.ErrorRedirectURL
	//if to != "" {
	//	redirectTo = to
	//}

	err := h.auditError(c, error)
	if err != nil {
		error = err
	}

	redirectURL := thirdparty.GetErrorUrl(to, error)
	return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

func (h *ThirdPartyHandler) auditError(c echo.Context, logError error) error {
	e, ok := logError.(*thirdparty.ThirdPartyError)

	tenant, err := context.GetTenant(c)
	if err != nil {
		return fmt.Errorf("failed to get tenant from context: %w", err)
	}

	var auditLogError error
	if ok && e.Code != thirdparty.ErrorCodeServerError {
		auditLogError = h.auditLogger.Create(c, models.AuditLogThirdPartySignInSignUpFailed, nil, logError, tenant.ID)
	} else {
		zeroLogger.Error().
			Str("time_unix", strconv.FormatInt(time.Now().Unix(), 10)).
			Str("id", c.Response().Header().Get(echo.HeaderXRequestID)).
			Str("remote_ip", c.RealIP()).
			Str("host", c.Request().Host).
			Str("method", c.Request().Method).
			Str("uri", c.Request().RequestURI).
			Str("user_agent", c.Request().UserAgent()).
			Err(logError).
			Send()
	}
	return auditLogError
}
