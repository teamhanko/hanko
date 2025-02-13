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
	samlUtils "github.com/teamhanko/hanko/backend/ee/saml/utils"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/session"
	"github.com/teamhanko/hanko/backend/thirdparty"
	"github.com/teamhanko/hanko/backend/utils"
	"net/http"
	"net/url"
	"strings"
	"time"
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

func (handler *Handler) callbackPostIdPInitiated(c echo.Context, samlResponse string) error {
	// ignore URL parse error because config validation already ensures it is a parseable URL
	redirectTo, _ := url.Parse(handler.samlService.Config().Saml.DefaultRedirectUrl)

	// We need to already parse the response to be able to extract information (a response's ID, Issuer, InResponseTo
	// nodes/values) to ensure protection against replaying IDP initiated responses as well as using service provider
	// issued responses as IDP initiated responses, even though we later also use the gosaml2 library to parse (and then
	// also validate) the response _again_. The reason is that the gosaml2 library does not make this information
	// easily/publicly accessible through its API.
	parsedSamlResponseDocument, _, err := samlUtils.ParseSamlResponse(samlResponse)
	if err != nil {
		return handler.redirectError(
			c,
			thirdparty.ErrorInvalidRequest("could not parse saml response").WithCause(err),
			redirectTo.String(),
		)
	}

	responseElement := parsedSamlResponseDocument.FindElement("/Response")
	if responseElement == nil {
		return handler.redirectError(
			c,
			thirdparty.ErrorInvalidRequest("invalid saml response: no response node present"),
			redirectTo.String(),
		)
	}

	issuerElement := parsedSamlResponseDocument.FindElement("/Response/Issuer")
	if issuerElement == nil || issuerElement.Text() == "" {
		return handler.redirectError(
			c,
			thirdparty.ErrorInvalidRequest("invalid saml response: no issuer node present"),
			redirectTo.String(),
		)
	}

	issuer := issuerElement.Text()

	serviceProvider, err := handler.samlService.GetProviderByIssuer(issuer)
	if err != nil {
		return handler.redirectError(
			c,
			thirdparty.ErrorInvalidRequest(
				fmt.Sprintf("could not get provider for issuer %s", issuer)).
				WithCause(err),
			redirectTo.String(),
		)
	}

	// We need to check whether this is an unsolicited request, otherwise SP initiated responses could
	// be used as IDP initiated responses.
	if responseElement.SelectAttr("InResponseTo") != nil {
		return handler.redirectError(
			c,
			thirdparty.ErrorInvalidRequest("saml request is not unsolicited"),
			redirectTo.String(),
		)
	}

	assertionInfo, err := handler.getAssertionInfo(serviceProvider, samlResponse)
	if err != nil {
		return handler.redirectError(
			c,
			thirdparty.ErrorInvalidRequest("could not get assertion info").WithCause(err),
			redirectTo.String(),
		)
	}

	samlResponseIDAttr := responseElement.SelectAttr("ID")
	if samlResponseIDAttr == nil {
		return handler.redirectError(
			c,
			thirdparty.ErrorInvalidRequest("invalid saml response: no ID for response present"),
			redirectTo.String(),
		)
	}

	samlResponseID := samlResponseIDAttr.Value

	samlIDPInitiatedRequestPersister := handler.samlService.Persister().GetSamlIDPInitiatedRequestPersister()

	// We use the SAML response's ID to prevent replay attacks by persisting every IDP initiated request and
	// checking whether an IDP initiated request already exists for this request.
	existingSamlIDPInitiatedRequest, err := samlIDPInitiatedRequestPersister.GetByResponseIDAndIssuer(samlResponseID, issuer)
	if existingSamlIDPInitiatedRequest != nil {
		return handler.redirectError(
			c,
			thirdparty.ErrorInvalidRequest("attempting to replay unsolicited saml request"),
			redirectTo.String(),
		)
	}

	// We assume only one assertion, and we assume it is present because we already validated it using the gosaml2
	// library (which also consumes only one/the first assertion). We also assume assertion conditions are present
	// because validation assures it is not nil (or else it returns an error).
	expiresAtString := assertionInfo.Assertions[0].Conditions.NotOnOrAfter

	expiresAt, err := time.Parse(time.RFC3339, expiresAtString)
	if err != nil {
		return handler.redirectError(
			c,
			thirdparty.ErrorServer("could not parse saml assertion conditions' NotOnOrAfter value").WithCause(err),
			redirectTo.String(),
		)
	}

	// If no request exists we create a new IDP initiated request model and persist it.
	samlIDPInitiatedRequest, err := models.NewSamlIDPInitiatedRequest(samlResponseID, issuer, expiresAt)
	if err != nil {
		return handler.redirectError(
			c,
			thirdparty.ErrorServer("could not instantiate saml idp initiated request model").WithCause(err),
			redirectTo.String(),
		)
	}

	err = samlIDPInitiatedRequestPersister.Create(*samlIDPInitiatedRequest)
	if err != nil {
		return handler.redirectError(
			c,
			thirdparty.ErrorServer("could not persist saml idp initiated request"),
			redirectTo.String(),
		)
	}

	redirectUrl, samlError := handler.linkAccount(c, redirectTo, true, serviceProvider, assertionInfo)
	if samlError != nil {
		return handler.redirectError(
			c,
			samlError,
			redirectTo.String(),
		)
	}

	// Add hint to the redirect URL that this is an IDP initiated request so that a token exchange can
	// eventually be performed through the dedicated flow API handler.
	values := redirectUrl.Query()
	values.Add("saml_hint", "idp_initiated")
	redirectUrl.RawQuery = values.Encode()

	return c.Redirect(http.StatusFound, redirectUrl.String())
}

func (handler *Handler) CallbackPost(c echo.Context) error {
	relayState := c.FormValue("RelayState")
	samlResponse := c.FormValue("SAMLResponse")

	if handler.isIDPInitiated(relayState) {
		return handler.callbackPostIdPInitiated(c, samlResponse)
	} else {
		state, err := VerifyState(
			handler.samlService.Config(),
			handler.samlService.Persister().GetSamlStatePersister(),
			strings.TrimPrefix(relayState, statePrefixServiceProviderInitiated),
		)

		if err != nil {
			return handler.redirectError(
				c,
				thirdparty.ErrorInvalidRequest(err.Error()).WithCause(err),
				handler.samlService.Config().Saml.DefaultRedirectUrl,
			)
		}

		if strings.TrimSpace(state.RedirectTo) == "" {
			state.RedirectTo = handler.samlService.Config().Saml.DefaultRedirectUrl
		}

		redirectTo, err := url.Parse(state.RedirectTo)
		if err != nil {
			return handler.redirectError(
				c,
				thirdparty.ErrorServer("unable to parse redirect url").WithCause(err),
				handler.samlService.Config().Saml.DefaultRedirectUrl,
			)
		}

		foundProvider, err := handler.samlService.GetProviderByDomain(state.Provider)
		if err != nil {
			return handler.redirectError(
				c,
				thirdparty.ErrorServer("unable to find provider by domain").WithCause(err),
				redirectTo.String(),
			)
		}

		assertionInfo, err := handler.getAssertionInfo(foundProvider, samlResponse)
		if err != nil {
			return handler.redirectError(
				c,
				thirdparty.ErrorServer("unable to parse saml response").WithCause(err),
				redirectTo.String(),
			)
		}

		redirectUrl, err := handler.linkAccount(c, redirectTo, state.IsFlow, foundProvider, assertionInfo)
		if err != nil {
			return handler.redirectError(
				c,
				err,
				redirectTo.String(),
			)
		}

		return c.Redirect(http.StatusFound, redirectUrl.String())
	}
}

func (handler *Handler) isIDPInitiated(relayState string) bool {
	return !strings.HasPrefix(relayState, statePrefixServiceProviderInitiated)
}

func (handler *Handler) linkAccount(c echo.Context, redirectTo *url.URL, isFlow bool, provider provider.ServiceProvider, assertionInfo *saml2.AssertionInfo) (*url.URL, error) {
	var accountLinkingResult *thirdparty.AccountLinkingResult
	var err error
	err = handler.samlService.Persister().Transaction(func(tx *pop.Connection) error {
		userdata := provider.GetUserData(assertionInfo)
		identityProviderIssuer := assertionInfo.Assertions[0].Issuer
		samlDomain := provider.GetDomain()
		linkResult, errTx := thirdparty.LinkAccount(tx, handler.samlService.Config(), handler.samlService.Persister(), userdata, identityProviderIssuer.Value, true, &samlDomain, isFlow)
		if errTx != nil {
			return errTx
		}

		accountLinkingResult = linkResult

		emailModel := linkResult.User.Emails.GetEmailByAddress(userdata.Metadata.Email)
		identityModel := emailModel.Identities.GetIdentity(identityProviderIssuer.Value, userdata.Metadata.Subject)

		token, errTx := models.NewToken(
			linkResult.User.ID,
			models.TokenWithIdentityID(identityModel.ID),
			models.TokenForFlowAPI(isFlow),
			models.TokenUserCreated(linkResult.UserCreated))
		if errTx != nil {
			return thirdparty.ErrorServer("could not create token").WithCause(errTx)
		}

		errTx = handler.samlService.Persister().GetTokenPersisterWithConnection(tx).Create(*token)
		if errTx != nil {
			return thirdparty.ErrorServer("could not save token to db").WithCause(errTx)
		}

		query := redirectTo.Query()
		query.Add(utils.HankoTokenQuery, token.Value)
		redirectTo.RawQuery = query.Encode()

		return nil
	})

	if err != nil {
		return nil, err
	}

	err = handler.auditLogger.Create(c, accountLinkingResult.Type, accountLinkingResult.User, nil)

	if err != nil {
		return nil, err
	}

	return redirectTo, nil
}

func (handler *Handler) getAssertionInfo(provider provider.ServiceProvider, samlResponse string) (*saml2.AssertionInfo, error) {
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
