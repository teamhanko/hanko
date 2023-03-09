package handler

import (
	"fmt"
	"github.com/labstack/echo/v4"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/crypto/jwk"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/session"
	"github.com/teamhanko/hanko/backend/thirdparty"
	"golang.org/x/oauth2"
	"net/http"
)

type ThirdPartyHandler struct {
	auditLogger    auditlog.Logger
	cfg            *config.Config
	persister      persistence.Persister
	sessionManager session.Manager
	jwkManager     jwk.Manager
}

func NewThirdPartyHandler(cfg *config.Config, persister persistence.Persister, sessionManager session.Manager, auditLogger auditlog.Logger, jwkManager jwk.Manager) *ThirdPartyHandler {
	return &ThirdPartyHandler{
		auditLogger:    auditLogger,
		cfg:            cfg,
		persister:      persister,
		sessionManager: sessionManager,
		jwkManager:     jwkManager,
	}
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

	state, err := thirdparty.GenerateState(h.cfg.ThirdParty, h.jwkManager, provider.Name(), request.RedirectTo)
	if err != nil {
		return h.redirectError(c, thirdparty.ErrorServer("could not generate state").WithCause(err), errorRedirectTo)
	}

	authCodeUrl := provider.AuthCodeURL(string(state), oauth2.SetAuthURLParam("prompt", "consent"))

	return c.Redirect(http.StatusTemporaryRedirect, authCodeUrl)
}

func (h *ThirdPartyHandler) Callback(c echo.Context) error {
	var callback dto.ThirdPartyAuthCallback
	err := c.Bind(&callback)
	if err != nil {
		return h.redirectError(c, thirdparty.ErrorServer("could not decode request payload").WithCause(err), h.cfg.ThirdParty.ErrorRedirectURL)
	}

	err = c.Validate(callback)
	if err != nil {
		return h.redirectError(c, thirdparty.ErrorInvalidRequest(err.Error()).WithCause(err), h.cfg.ThirdParty.ErrorRedirectURL)
	}

	state, err := thirdparty.VerifyState(h.sessionManager, callback.State)
	if err != nil {
		return h.redirectError(c, thirdparty.ErrorInvalidRequest(err.Error()).WithCause(err), h.cfg.ThirdParty.ErrorRedirectURL)
	}

	if callback.HasError() {
		return h.redirectError(c, thirdparty.NewThirdPartyError(callback.Error, callback.ErrorDescription), h.cfg.ThirdParty.ErrorRedirectURL)
	}

	provider, err := thirdparty.GetProvider(h.cfg.ThirdParty, state.Provider)
	if err != nil {
		return h.redirectError(c, thirdparty.ErrorInvalidRequest(err.Error()).WithCause(err), h.cfg.ThirdParty.ErrorRedirectURL)
	}

	if callback.AuthCode == "" {
		return h.redirectError(c, thirdparty.ErrorInvalidRequest("auth code missing from request"), h.cfg.ThirdParty.ErrorRedirectURL)
	}

	token, err := provider.GetOAuthToken(callback.AuthCode)
	if err != nil {
		return h.redirectError(c, thirdparty.ErrorInvalidRequest("could not exchange authorization code for access token").WithCause(err), h.cfg.ThirdParty.ErrorRedirectURL)
	}

	userData, err := provider.GetUserData(token)
	if err != nil {
		return h.redirectError(c, thirdparty.ErrorInvalidRequest("could not retrieve user data from provider").WithCause(err), h.cfg.ThirdParty.ErrorRedirectURL)
	}

	linkingResult, err := thirdparty.LinkAccount(h.cfg, h.persister, userData, provider.Name())
	if err != nil {
		return h.redirectError(c, err, h.cfg.ThirdParty.ErrorRedirectURL)
	}

	jwt, err := h.sessionManager.GenerateJWT(linkingResult.User.ID)
	if err != nil {
		return h.redirectError(c, thirdparty.ErrorServer("could not generate jwt").WithCause(err), h.cfg.ThirdParty.ErrorRedirectURL)
	}

	cookie, err := h.sessionManager.GenerateCookie(jwt)
	if err != nil {
		return h.redirectError(c, thirdparty.ErrorServer("could not create session cookie").WithCause(err), h.cfg.ThirdParty.ErrorRedirectURL)
	}

	c.SetCookie(cookie)

	if h.cfg.Session.EnableAuthTokenHeader {
		c.Response().Header().Set("X-Auth-Token", jwt)
	}

	err = h.auditLogger.Create(c, linkingResult.Type, linkingResult.User, nil)
	if err != nil {
		return h.redirectError(c, thirdparty.ErrorServer("could not create audit log").WithCause(err), h.cfg.ThirdParty.ErrorRedirectURL)
	}

	return c.Redirect(http.StatusTemporaryRedirect, state.RedirectTo)
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
	if ok && e.Code != thirdparty.ThirdPartyErrorCodeServerError {
		auditLogError = h.auditLogger.Create(c, models.AuditLogThirdPartySignInSignUpFailed, nil, err)
	}
	return auditLogError
}
