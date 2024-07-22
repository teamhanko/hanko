package saml

import (
	"errors"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/labstack/echo/v4"
	saml2 "github.com/russellhaering/gosaml2"
	auditlog "github.com/teamhanko/hanko/backend/audit_log"
	"github.com/teamhanko/hanko/backend/ee/saml/dto"
	"github.com/teamhanko/hanko/backend/ee/saml/provider"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/session"
	"github.com/teamhanko/hanko/backend/thirdparty"
	"github.com/teamhanko/hanko/backend/utils"
	"net/http"
	"net/url"
	"strings"
)

type Handler struct {
	auditLogger    auditlog.Logger
	sessionManager session.Manager
	samlService    Service
}

func NewSamlHandler(sessionManager session.Manager, auditLogger auditlog.Logger, samlService Service) *Handler {
	return &Handler{
		auditLogger:    auditLogger,
		sessionManager: sessionManager,
		samlService:    samlService,
	}
}

func (handler *Handler) Metadata(c echo.Context) error {
	var request dto.SamlMetadataRequest
	err := c.Bind(&request)
	if err != nil {
		return c.JSON(http.StatusBadRequest, thirdparty.ErrorInvalidRequest("domain is missing"))
	}

	foundProvider, err := handler.samlService.GetProviderByDomain(request.Domain)
	if err != nil {
		return c.NoContent(http.StatusNotFound)
	}

	if request.CertOnly {
		cert, err := handler.samlService.Persister().GetSamlCertificatePersister().GetFirst()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, thirdparty.ErrorServer("unable to provide metadata").WithCause(err))
		}

		if cert == nil {
			return c.NoContent(http.StatusNotFound)
		}

		c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%s-service-provider.pem", handler.samlService.Config().Service.Name))
		return c.Blob(http.StatusOK, echo.MIMEOctetStream, []byte(cert.CertData))
	}

	xmlMetadata, err := foundProvider.ProvideMetadataAsXml()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, thirdparty.ErrorServer("unable to provide metadata").WithCause(err))
	}

	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%s-metadata.xml", handler.samlService.Config().Service.Name))
	return c.Blob(http.StatusOK, echo.MIMEOctetStream, xmlMetadata)
}

func (handler *Handler) Auth(c echo.Context) error {
	errorRedirectTo := c.Request().Header.Get("Referer")
	if errorRedirectTo == "" {
		errorRedirectTo = handler.samlService.Config().Saml.DefaultRedirectUrl
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

	foundProvider, err := handler.samlService.GetProviderByDomain(request.Domain)
	if err != nil {
		return handler.redirectError(c, thirdparty.ErrorInvalidRequest(err.Error()).WithCause(err), errorRedirectTo)
	}

	redirectUrl, err := handler.samlService.GetAuthUrl(foundProvider, request.RedirectTo, false)
	if err != nil {
		return handler.redirectError(c, thirdparty.ErrorServer("could not generate auth url").WithCause(err), errorRedirectTo)
	}

	return c.Redirect(http.StatusTemporaryRedirect, redirectUrl)
}

func (handler *Handler) CallbackPost(c echo.Context) error {
	state, samlError := VerifyState(handler.samlService.Config(), handler.samlService.Persister().GetSamlStatePersister(), c.FormValue("RelayState"))
	if samlError != nil {
		return handler.redirectError(
			c,
			thirdparty.ErrorInvalidRequest(samlError.Error()).WithCause(samlError),
			handler.samlService.Config().Saml.DefaultRedirectUrl,
		)
	}

	if strings.TrimSpace(state.RedirectTo) == "" {
		state.RedirectTo = handler.samlService.Config().Saml.DefaultRedirectUrl
	}

	redirectTo, samlError := url.Parse(state.RedirectTo)
	if samlError != nil {
		return handler.redirectError(
			c,
			thirdparty.ErrorServer("unable to parse redirect url").WithCause(samlError),
			handler.samlService.Config().Saml.DefaultRedirectUrl,
		)
	}

	foundProvider, samlError := handler.samlService.GetProviderByDomain(state.Provider)
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

func (handler *Handler) linkAccount(c echo.Context, redirectTo *url.URL, state *State, provider provider.ServiceProvider, assertionInfo *saml2.AssertionInfo) (*url.URL, error) {
	var accountLinkingResult *thirdparty.AccountLinkingResult
	var samlError error
	samlError = handler.samlService.Persister().Transaction(func(tx *pop.Connection) error {
		userdata := provider.GetUserData(assertionInfo)

		linkResult, samlErrorTx := thirdparty.LinkAccount(tx, handler.samlService.Config(), handler.samlService.Persister(), userdata, state.Provider, true, state.IsFlow)
		if samlErrorTx != nil {
			return samlErrorTx
		}
		accountLinkingResult = linkResult

		emailModel := linkResult.User.Emails.GetEmailByAddress(userdata.Metadata.Email)
		identityModel := emailModel.Identities.GetIdentity(provider.GetDomain(), userdata.Metadata.Subject)

		token, tokenError := models.NewToken(
			linkResult.User.ID,
			models.TokenWithIdentityID(identityModel.ID),
			models.TokenForFlowAPI(state.IsFlow),
			models.TokenUserCreated(linkResult.UserCreated))
		if tokenError != nil {
			return thirdparty.ErrorServer("could not create token").WithCause(tokenError)
		}

		tokenError = handler.samlService.Persister().GetTokenPersisterWithConnection(tx).Create(*token)
		if tokenError != nil {
			return thirdparty.ErrorServer("could not save token to db").WithCause(tokenError)
		}

		query := redirectTo.Query()
		query.Add(utils.HankoTokenQuery, token.Value)
		redirectTo.RawQuery = query.Encode()

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

func (handler *Handler) parseSamlResponse(provider provider.ServiceProvider, samlResponse string) (*saml2.AssertionInfo, error) {
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

func (handler *Handler) redirectError(c echo.Context, error error, to string) error {
	c.Logger().Error(error)

	err := handler.auditError(c, error)
	if err != nil {
		error = err
	}

	redirectURL := thirdparty.GetErrorUrl(to, error)
	return c.Redirect(http.StatusSeeOther, redirectURL)
}

func (handler *Handler) auditError(c echo.Context, err error) error {
	var e *thirdparty.ThirdPartyError
	ok := errors.As(err, &e)

	var auditLogError error
	if ok && e.Code != thirdparty.ErrorCodeServerError {
		auditLogError = handler.auditLogger.Create(c, models.AuditLogThirdPartySignInSignUpFailed, nil, err)
	}
	return auditLogError
}

func (handler *Handler) GetProvider(c echo.Context) error {
	var request dto.SamlRequest
	err := c.Bind(&request)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	foundProvider, err := handler.samlService.GetProviderByDomain(request.Domain)
	if err != nil {
		return c.NoContent(http.StatusNotFound)
	}

	return c.JSON(http.StatusOK, foundProvider.GetConfig())
}
