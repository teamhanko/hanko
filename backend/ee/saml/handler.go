package saml

import (
	"errors"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/labstack/echo/v4"
	saml2 "github.com/russellhaering/gosaml2"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/ee/saml/dto"
	"github.com/teamhanko/hanko/backend/ee/saml/provider"
	samlUtils "github.com/teamhanko/hanko/backend/ee/saml/utils"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/session"
	"github.com/teamhanko/hanko/backend/thirdparty"
	"github.com/teamhanko/hanko/backend/utils"
	"net/http"
	"net/url"
	"strings"
)

type SamlHandler struct {
	auditLogger    auditlog.Logger
	config         *config.Config
	persister      persistence.Persister
	sessionManager session.Manager
	providers      []provider.ServiceProvider
}

func NewSamlHandler(cfg *config.Config, persister persistence.Persister, sessionManager session.Manager, auditLogger auditlog.Logger) *SamlHandler {
	providers := make([]provider.ServiceProvider, 0)
	for _, idpConfig := range cfg.Saml.IdentityProviders {
		if idpConfig.Enabled {
			hostName := ""
			hostName, err := parseProviderFromMetadataUrl(idpConfig.MetadataUrl)
			if err != nil {
				fmt.Printf("failed to parse provider '%s' from metadata url: %v\n", idpConfig.Name, err)
				continue
			}

			newProvider, err := provider.GetProvider(hostName, cfg, idpConfig, persister.GetSamlCertificatePersister())
			if err != nil {
				fmt.Printf("failed to initialize provider '%s': %v\n", idpConfig.Name, err)
				continue
			}

			providers = append(providers, newProvider)
		}
	}

	return &SamlHandler{
		auditLogger:    auditLogger,
		config:         cfg,
		persister:      persister,
		sessionManager: sessionManager,
		providers:      providers,
	}
}

func parseProviderFromMetadataUrl(idpUrlString string) (string, error) {
	idpUrl, err := url.Parse(idpUrlString)
	if err != nil {
		return "", err
	}

	return idpUrl.Host, nil
}

func (handler *SamlHandler) getProviderByDomain(domain string) (provider.ServiceProvider, error) {
	for _, availableProvider := range handler.providers {
		if availableProvider.GetDomain() == domain {
			return availableProvider, nil
		}
	}

	return nil, fmt.Errorf("unknown provider for domain %s", domain)
}

func (handler *SamlHandler) Metadata(c echo.Context) error {
	var request dto.SamlMetadataRequest
	err := c.Bind(&request)
	if err != nil {
		return c.JSON(http.StatusBadRequest, thirdparty.ErrorInvalidRequest("domain is missing"))
	}

	foundProvider, err := handler.getProviderByDomain(request.Domain)
	if err != nil {
		return c.NoContent(http.StatusNotFound)
	}

	if request.CertOnly {
		cert, err := handler.persister.GetSamlCertificatePersister().GetFirst()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, thirdparty.ErrorServer("unable to provide metadata").WithCause(err))
		}

		if cert == nil {
			return c.NoContent(http.StatusNotFound)
		}

		c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%s-service-provider.pem", handler.config.Service.Name))
		return c.Blob(http.StatusOK, echo.MIMEOctetStream, []byte(cert.CertData))
	}

	xmlMetadata, err := foundProvider.ProvideMetadataAsXml()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, thirdparty.ErrorServer("unable to provide metadata").WithCause(err))
	}

	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%s-metadata.xml", handler.config.Service.Name))
	return c.Blob(http.StatusOK, echo.MIMEOctetStream, xmlMetadata)
}

func (handler *SamlHandler) Auth(c echo.Context) error {
	errorRedirectTo := c.Request().Header.Get("Referer")
	if errorRedirectTo == "" {
		errorRedirectTo = handler.config.Saml.DefaultRedirectUrl
	}

	var request dto.SamlAuthRequest
	err := c.Bind(&request)
	if err != nil {
		return handler.redirectError(c, thirdparty.ErrorInvalidRequest(err.Error()).WithCause(err), errorRedirectTo)
	}

	err = c.Validate(request)
	if err != nil {
		return handler.redirectError(c, thirdparty.ErrorInvalidRequest(err.Error()).WithCause(err), errorRedirectTo)
	}

	if ok := samlUtils.IsAllowedRedirect(handler.config.Saml, request.RedirectTo); !ok {
		return handler.redirectError(c, thirdparty.ErrorInvalidRequest(fmt.Sprintf("redirect to '%s' not allowed", request.RedirectTo)), errorRedirectTo)
	}

	foundProvider, err := handler.getProviderByDomain(request.Domain)
	if err != nil {
		return handler.redirectError(c, thirdparty.ErrorInvalidRequest(err.Error()).WithCause(err), errorRedirectTo)
	}

	state, err := GenerateState(
		handler.config,
		handler.persister.GetSamlStatePersister(),
		request.Domain,
		request.RedirectTo)

	if err != nil {
		return handler.redirectError(c, thirdparty.ErrorServer("could not generate state").WithCause(err), errorRedirectTo)
	}

	redirectUrl, err := foundProvider.GetService().BuildAuthURL(string(state))
	if err != nil {
		return handler.redirectError(c, thirdparty.ErrorServer("could not generate auth url").WithCause(err), errorRedirectTo)
	}

	return c.Redirect(http.StatusTemporaryRedirect, redirectUrl)
}

func (handler *SamlHandler) CallbackPost(c echo.Context) error {
	state, samlError := VerifyState(handler.config, handler.persister.GetSamlStatePersister(), c.FormValue("RelayState"))
	if samlError != nil {
		return handler.redirectError(
			c,
			thirdparty.ErrorInvalidRequest(samlError.Error()).WithCause(samlError),
			handler.config.Saml.DefaultRedirectUrl,
		)
	}

	if strings.TrimSpace(state.RedirectTo) == "" {
		state.RedirectTo = handler.config.Saml.DefaultRedirectUrl
	}

	redirectTo, samlError := url.Parse(state.RedirectTo)
	if samlError != nil {
		return handler.redirectError(
			c,
			thirdparty.ErrorServer("unable to parse redirect url").WithCause(samlError),
			handler.config.Saml.DefaultRedirectUrl,
		)
	}

	foundProvider, samlError := handler.getProviderByDomain(state.Provider)
	if samlError != nil {
		return handler.redirectError(
			c,
			thirdparty.ErrorServer("unable to find provider by domain").WithCause(samlError),
			redirectTo.String(),
		)
	}

	assertionInfo, samlError := handler.parseSamlResponse(foundProvider, c.FormValue("SAMLResponse"))
	if samlError != nil {
		return handler.redirectError(
			c,
			thirdparty.ErrorServer("unable to parse saml response").WithCause(samlError),
			redirectTo.String(),
		)
	}

	redirectUrl, samlError := handler.linkAccount(c, redirectTo, state, foundProvider, assertionInfo)
	if samlError != nil {
		return handler.redirectError(
			c,
			samlError,
			redirectTo.String(),
		)
	}

	return c.Redirect(http.StatusFound, redirectUrl.String())
}

func (handler *SamlHandler) linkAccount(c echo.Context, redirectTo *url.URL, state *State, provider provider.ServiceProvider, assertionInfo *saml2.AssertionInfo) (*url.URL, error) {
	var accountLinkingResult *thirdparty.AccountLinkingResult
	var samlError error
	samlError = handler.persister.Transaction(func(tx *pop.Connection) error {
		userdata := provider.GetUserData(assertionInfo)

		linkResult, samlError := thirdparty.LinkAccount(tx, handler.config, handler.persister, userdata, state.Provider, true)
		if samlError != nil {
			return samlError
		}
		accountLinkingResult = linkResult

		token, samlError := handler.createHankoToken(linkResult, tx)
		if samlError != nil {
			return samlError
		}

		query := redirectTo.Query()
		query.Add(utils.HankoTokenQuery, token.Value)
		redirectTo.RawQuery = query.Encode()

		cookie := utils.GenerateStateCookie(handler.config, utils.HankoThirdpartyStateCookie, "", utils.CookieOptions{
			MaxAge:   -1,
			Path:     "/",
			SameSite: http.SameSiteLaxMode,
		})
		c.SetCookie(cookie)

		return nil

	})

	if samlError != nil {
		return nil, samlError
	}

	samlError = handler.auditLogger.Create(c, accountLinkingResult.Type, accountLinkingResult.User, nil)

	if samlError != nil {
		return nil, samlError
	}

	return redirectTo, nil
}

func (handler *SamlHandler) createHankoToken(linkResult *thirdparty.AccountLinkingResult, tx *pop.Connection) (*models.Token, error) {
	token, tokenError := models.NewToken(linkResult.User.ID)
	if tokenError != nil {
		return nil, thirdparty.ErrorServer("could not create token").WithCause(tokenError)
	}

	tokenError = handler.persister.GetTokenPersisterWithConnection(tx).Create(*token)
	if tokenError != nil {
		return nil, thirdparty.ErrorServer("could not save token to db").WithCause(tokenError)
	}

	return token, nil
}

func (handler *SamlHandler) parseSamlResponse(provider provider.ServiceProvider, samlResponse string) (*saml2.AssertionInfo, error) {
	assertionInfo, err := provider.GetService().RetrieveAssertionInfo(samlResponse)
	if err != nil {
		return nil, thirdparty.ErrorServer("unable to parse SAML response").WithCause(err)
	}

	if assertionInfo.WarningInfo.InvalidTime {
		return nil, thirdparty.ErrorServer("SAMLAssertion expired")
	}

	if assertionInfo.WarningInfo.NotInAudience {
		return nil, thirdparty.ErrorServer("not in SAML audience")
	}

	return assertionInfo, nil
}

func (handler *SamlHandler) redirectError(c echo.Context, error error, to string) error {
	err := handler.auditError(c, error)
	if err != nil {
		error = err
	}

	redirectURL := thirdparty.GetErrorUrl(to, error)
	return c.Redirect(http.StatusSeeOther, redirectURL)
}

func (handler *SamlHandler) auditError(c echo.Context, err error) error {
	var e *thirdparty.ThirdPartyError
	ok := errors.As(err, &e)

	var auditLogError error
	if ok && e.Code != thirdparty.ErrorCodeServerError {
		auditLogError = handler.auditLogger.Create(c, models.AuditLogThirdPartySignInSignUpFailed, nil, err)
	}
	return auditLogError
}

func (handler *SamlHandler) GetProvider(c echo.Context) error {
	var request dto.SamlRequest
	err := c.Bind(&request)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	foundProvider, err := handler.getProviderByDomain(request.Domain)
	if err != nil {
		return c.NoContent(http.StatusNotFound)
	}

	return c.JSON(http.StatusOK, foundProvider.GetConfig())
}
