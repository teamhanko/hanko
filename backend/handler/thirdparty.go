package handler

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gobuffalo/pop/v6"
	"github.com/labstack/echo/v4"
	auditlog "github.com/teamhanko/hanko/backend/v2/audit_log"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/dto"
	"github.com/teamhanko/hanko/backend/v2/dto/admin"
	"github.com/teamhanko/hanko/backend/v2/persistence"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
	"github.com/teamhanko/hanko/backend/v2/thirdparty"
	"github.com/teamhanko/hanko/backend/v2/utils"
	webhookUtils "github.com/teamhanko/hanko/backend/v2/webhooks/utils"
	"golang.org/x/oauth2"
)

type ThirdPartyHandler struct {
	auditLogger auditlog.Logger
	persister   persistence.Persister
}

func NewThirdPartyHandler(persister persistence.Persister, auditLogger auditlog.Logger) *ThirdPartyHandler {
	return &ThirdPartyHandler{
		auditLogger: auditLogger,
		persister:   persister,
	}
}

func (h *ThirdPartyHandler) CallbackPost(c echo.Context) error {
	tenantConfig := c.Get("tenant_config").(*config.TenantConfig)
	if tenantConfig == nil {
		return h.redirectError(c, thirdparty.ErrorServer("could not get tenant config"), "/error") // TODO:
	}
	q, err := c.FormParams()
	if err != nil {
		return h.redirectError(c, thirdparty.ErrorServer("could not get form parameters"), tenantConfig.ThirdParty.ErrorRedirectURL)
	}
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/thirdparty/callback?%s", q.Encode()))
}

func (h *ThirdPartyHandler) Callback(c echo.Context) error {
	tenantConfig := c.Get("tenant_config").(*config.TenantConfig)
	if tenantConfig == nil {
		return h.redirectError(c, thirdparty.ErrorServer("could not get tenant config"), "/error") // TODO:
	}
	tenantID, err := utils.TenantIDFromContext(c)
	if err != nil {
		return fmt.Errorf("invalid tenant identifier: %w", err)
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
		state, terr = thirdparty.VerifyState(tenantConfig, callback.State, expectedState)
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

		provider, terr := thirdparty.GetProvider(tenantConfig.ThirdParty, state.Provider)
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
		oAuthToken, terr := provider.GetOAuthToken(callback.AuthCode, opts...)
		if terr != nil {
			return thirdparty.ErrorInvalidRequest("could not exchange authorization code for access token").WithCause(terr)
		}

		userData, terr := provider.GetUserData(oAuthToken)
		if terr != nil {
			return thirdparty.ErrorInvalidRequest("could not retrieve user data from provider").WithCause(terr)
		}

		linkingResult, terr := thirdparty.LinkAccount(tx, tenantConfig, h.persister, userData, provider.ID(), false, nil, state.IsFlow, state.UserID, tenantID)
		if terr != nil {
			return terr
		}
		accountLinkingResult = linkingResult

		identityModel, err := h.persister.GetIdentityPersisterWithConnection(tx).Get(userData.Metadata.Subject, provider.ID(), tenantID)
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
			Domain:   tenantConfig.Session.Cookie.Domain,
			MaxAge:   -1,
			Secure:   tenantConfig.Session.Cookie.Secure,
			HttpOnly: tenantConfig.Session.Cookie.HttpOnly,
			SameSite: http.SameSiteLaxMode,
		})

		return nil
	})

	errorRedirect := tenantConfig.ThirdParty.ErrorRedirectURL
	if redirectToURL != nil {
		errorRedirect = redirectToURL.String()
	}

	if err != nil {
		return h.redirectError(c, err, errorRedirect)
	}

	err = h.auditLogger.Create(c, accountLinkingResult.Type, accountLinkingResult.User, nil, tenantID)
	if err != nil {
		return h.redirectError(c, thirdparty.ErrorServer("could not create audit log").WithCause(err), errorRedirect)
	}

	if accountLinkingResult.WebhookEvent != nil {
		err = webhookUtils.TriggerWebhooks(c, h.persister.GetConnection(), tenantID, *accountLinkingResult.WebhookEvent, admin.FromUserModel(*accountLinkingResult.User))
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

func (h *ThirdPartyHandler) auditError(c echo.Context, err error) error {
	e, ok := err.(*thirdparty.ThirdPartyError)

	tenantID, err := utils.TenantIDFromContext(c)
	if err != nil {
		return fmt.Errorf("invalid tenant identifier: %w", err)
	}

	var auditLogError error
	if ok && e.Code != thirdparty.ErrorCodeServerError {
		auditLogError = h.auditLogger.Create(c, models.AuditLogThirdPartySignInSignUpFailed, nil, err, tenantID)
	}
	return auditLogError
}
