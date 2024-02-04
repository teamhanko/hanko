package handler

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"

	oidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/gobuffalo/pop/v6"
	"github.com/labstack/echo/v4"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/session"
	"github.com/teamhanko/hanko/backend/thirdparty"
	"github.com/teamhanko/hanko/backend/utils"
	"golang.org/x/oauth2"
)

type ThirdPartyHandler struct {
	auditLogger    auditlog.Logger
	cfg            *config.Config
	persister      persistence.Persister
	sessionManager session.Manager
}

func NewThirdPartyHandler(cfg *config.Config, persister persistence.Persister, sessionManager session.Manager, auditLogger auditlog.Logger) *ThirdPartyHandler {
	return &ThirdPartyHandler{
		auditLogger:    auditLogger,
		cfg:            cfg,
		persister:      persister,
		sessionManager: sessionManager,
	}
}
func (h *ThirdPartyHandler) randString(nByte int) (string, error) {
	b := make([]byte, nByte)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (h *ThirdPartyHandler) AuthOAuth(c echo.Context, oauthProvider thirdparty.OAuthProvider) error {
	errorRedirectTo := c.Request().Header.Get("Referer")
	if errorRedirectTo == "" {
		errorRedirectTo = h.cfg.ThirdParty.ErrorRedirectURL
	}

	var request dto.ThirdPartyAuthRequest
	err := c.Bind(&request)
	if err != nil {
		return h.redirectError(c, thirdparty.ErrorServer("could not decode request payload").WithCause(err), errorRedirectTo)
	}

	state, err := thirdparty.GenerateState(h.cfg, oauthProvider.Name(), request.RedirectTo)
	if err != nil {
		return h.redirectError(c, thirdparty.ErrorServer("could not generate state").WithCause(err), errorRedirectTo)
	}
	authCodeOptions := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("prompt", "consent"),
	}
	if oauthProvider.RequireNonce() {
		nonce, err := h.randString(16)
		if err != nil {
			return h.redirectError(c, thirdparty.ErrorServer("internal error").WithCause(err), errorRedirectTo)
		}
		nonceCookie := utils.GenerateStateCookie(h.cfg,
			utils.HankoThirdpartyNonceCookie, string(nonce), utils.CookieOptions{
				MaxAge:   300,
				Path:     "/",
				SameSite: http.SameSiteLaxMode,
			})
		c.SetCookie(nonceCookie)
		authCodeOptions = append(authCodeOptions, oidc.Nonce(nonce))
	}
	authCodeUrl := oauthProvider.AuthCodeURL(string(state), authCodeOptions...)

	cookie := utils.GenerateStateCookie(h.cfg,
		utils.HankoThirdpartyStateCookie, string(state), utils.CookieOptions{
			MaxAge:   300,
			Path:     "/",
			SameSite: http.SameSiteLaxMode,
		})

	c.SetCookie(cookie)

	return c.Redirect(http.StatusTemporaryRedirect, authCodeUrl)

}

func (h *ThirdPartyHandler) Auth(c echo.Context) error {
	errorRedirectTo := c.Request().Header.Get("Referer")
	if errorRedirectTo == "" {
		errorRedirectTo = h.cfg.ThirdParty.ErrorRedirectURL
	}

	var request dto.ThirdPartyAuthRequest
	err := c.Bind(&request)
	if err != nil {
		return h.redirectError(c, thirdparty.ErrorServer("could not decode request payload").WithCause(err), errorRedirectTo)
	}

	err = c.Validate(request)
	if err != nil {
		return h.redirectError(c, thirdparty.ErrorInvalidRequest(err.Error()).WithCause(err), errorRedirectTo)
	}

	if ok := thirdparty.IsAllowedRedirect(h.cfg.ThirdParty, request.RedirectTo); !ok {
		return h.redirectError(c, thirdparty.ErrorInvalidRequest(fmt.Sprintf("redirect to '%s' not allowed", request.RedirectTo)), errorRedirectTo)
	}

	provider, err := thirdparty.GetProvider(h.cfg.ThirdParty, request.Provider)
	if err != nil {
		return h.redirectError(c, thirdparty.ErrorInvalidRequest(err.Error()).WithCause(err), errorRedirectTo)
	}
	if provider.OAuthProvider != nil {
		return h.AuthOAuth(c, provider.OAuthProvider)
	}
	return h.redirectError(c, thirdparty.ErrorInvalidRequest(fmt.Sprintf("provider '%s' is not supported", request.Provider)), errorRedirectTo)

}

func (h *ThirdPartyHandler) CallbackPost(c echo.Context) error {
	q, err := c.FormParams()
	if err != nil {
		return h.redirectError(c, thirdparty.ErrorServer("could not get form parameters"), h.cfg.ThirdParty.ErrorRedirectURL)
	}
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/thirdparty/callback?%s", q.Encode()))
}

func (h *ThirdPartyHandler) Callback(c echo.Context) error {
	var successRedirectTo *url.URL
	var accountLinkingResult *thirdparty.AccountLinkingResult
	err := h.persister.Transaction(func(tx *pop.Connection) error {
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

		expectedState, terr := c.Cookie(utils.HankoThirdpartyStateCookie)
		if terr != nil {
			return thirdparty.ErrorInvalidRequest("thirdparty state cookie is missing")
		}

		state, terr := thirdparty.VerifyState(h.cfg, callback.State, expectedState.Value)
		if terr != nil {
			return thirdparty.ErrorInvalidRequest(terr.Error()).WithCause(terr)
		}

		if callback.HasError() {
			return thirdparty.NewThirdPartyError(callback.Error, callback.ErrorDescription)
		}

		provider, terr := thirdparty.GetProvider(h.cfg.ThirdParty, state.Provider)
		if terr != nil {
			return thirdparty.ErrorInvalidRequest(terr.Error()).WithCause(terr)
		}
		oauthProvider := provider.OAuthProvider
		if callback.AuthCode == "" {
			return thirdparty.ErrorInvalidRequest("auth code missing from request")
		}

		oAuthToken, terr := oauthProvider.GetOAuthToken(callback.AuthCode)
		if terr != nil {
			return thirdparty.ErrorInvalidRequest("could not exchange authorization code for access token").WithCause(terr)
		}

		userData, terr := oauthProvider.GetUserData(oAuthToken)
		if terr != nil {
			return thirdparty.ErrorInvalidRequest("could not retrieve user data from provider").WithCause(terr)
		}

		linkingResult, terr := thirdparty.LinkAccount(tx, h.cfg, h.persister, userData, oauthProvider.Name())
		if terr != nil {
			return terr
		}
		accountLinkingResult = linkingResult

		token, terr := models.NewToken(linkingResult.User.ID)
		if terr != nil {
			return thirdparty.ErrorServer("could not create token").WithCause(terr)
		}

		terr = h.persister.GetTokenPersisterWithConnection(tx).Create(*token)
		if terr != nil {
			return thirdparty.ErrorServer("could not save token to db").WithCause(terr)
		}

		redirectTo, terr := url.Parse(state.RedirectTo)
		if terr != nil {
			return thirdparty.ErrorServer("could not parse redirect url").WithCause(terr)
		}

		query := redirectTo.Query()
		query.Add(utils.HankoTokenQuery, token.Value)
		redirectTo.RawQuery = query.Encode()
		successRedirectTo = redirectTo

		c.SetCookie(&http.Cookie{
			Name:     utils.HankoThirdpartyStateCookie,
			Value:    "",
			Path:     "/",
			Domain:   h.cfg.Session.Cookie.Domain,
			MaxAge:   -1,
			Secure:   h.cfg.Session.Cookie.Secure,
			HttpOnly: h.cfg.Session.Cookie.HttpOnly,
			SameSite: http.SameSiteLaxMode,
		})

		return nil
	})

	if err != nil {
		return h.redirectError(c, err, h.cfg.ThirdParty.ErrorRedirectURL)
	}

	err = h.auditLogger.Create(c, accountLinkingResult.Type, accountLinkingResult.User, nil)

	if err != nil {
		return h.redirectError(c, thirdparty.ErrorServer("could not create audit log").WithCause(err), h.cfg.ThirdParty.ErrorRedirectURL)
	}

	return c.Redirect(http.StatusTemporaryRedirect, successRedirectTo.String())
}

func (h *ThirdPartyHandler) redirectError(c echo.Context, error error, to string) error {
	redirectTo := h.cfg.ThirdParty.ErrorRedirectURL
	if to != "" {
		redirectTo = to
	}

	err := h.auditError(c, error)
	if err != nil {
		error = err
	}

	redirectURL := thirdparty.GetErrorUrl(redirectTo, error)
	return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

func (h *ThirdPartyHandler) auditError(c echo.Context, err error) error {
	e, ok := err.(*thirdparty.ThirdPartyError)

	var auditLogError error
	if ok && e.Code != thirdparty.ErrorCodeServerError {
		auditLogError = h.auditLogger.Create(c, models.AuditLogThirdPartySignInSignUpFailed, nil, err)
	}
	return auditLogError
}
