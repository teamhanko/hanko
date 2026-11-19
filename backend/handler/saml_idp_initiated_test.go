package handler

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/beevik/etree"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	dsig "github.com/russellhaering/goxmldsig"
	"github.com/stretchr/testify/suite"
	auditlog "github.com/teamhanko/hanko/backend/v3/audit_log"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/context"
	"github.com/teamhanko/hanko/backend/v3/dto"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
	"github.com/teamhanko/hanko/backend/v3/saml"
	"github.com/teamhanko/hanko/backend/v3/test"
	"github.com/teamhanko/hanko/backend/v3/utils"
)

const (
	samlTestEndpoint   = "https://hanko.test.example/api"
	samlTestACSURL     = samlTestEndpoint + "/saml/callback"
	samlTestAudience   = "https://hanko.test.example"
	samlTestIdPIssuer  = "https://idp.test.example/metadata"
	samlTestEmailAttr  = "email"
	samlTestSubject    = "saml-user@test.example"
	samlTestRedirectTo = "https://app.test.example/login-complete"
)

func TestSamlIdPInitiatedSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(samlIdPInitiatedSuite))
}

type samlIdPInitiatedSuite struct {
	test.Suite
}

// samlTestIdP signs assertions but not the Response.
type samlTestIdP struct {
	key         *rsa.PrivateKey
	certDER     []byte
	metadataURL string
}

func (s *samlIdPInitiatedSuite) setUpIdP() (*samlTestIdP, func()) {
	s.T().Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	s.Require().NoError(err)

	now := time.Now().UTC()
	certDER, err := x509.CreateCertificate(rand.Reader, &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test-idp"},
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              now.Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}, &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test-idp"},
		NotBefore:    now.Add(-time.Hour),
		NotAfter:     now.Add(24 * time.Hour),
	}, &key.PublicKey, key)
	s.Require().NoError(err)

	metadata := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata" entityID="%s">
  <IDPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
    <KeyDescriptor use="signing">
      <KeyInfo xmlns="http://www.w3.org/2000/09/xmldsig#">
        <X509Data><X509Certificate>%s</X509Certificate></X509Data>
      </KeyInfo>
    </KeyDescriptor>
    <SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="https://idp.test.example/sso"/>
  </IDPSSODescriptor>
</EntityDescriptor>`, samlTestIdPIssuer, base64.StdEncoding.EncodeToString(certDER))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(metadata))
	}))

	return &samlTestIdP{key: key, certDER: certDER, metadataURL: server.URL}, server.Close
}

// The assertion declares both namespaces itself so its signature survives being nested in the Response.
func (s *samlIdPInitiatedSuite) samlResponse(idp *samlTestIdP, responseID, assertionID, inResponseTo string) string {
	s.T().Helper()

	now := time.Now().UTC()
	notOnOrAfter := now.Add(10 * time.Minute).Format(time.RFC3339)

	subjectConfirmationData := fmt.Sprintf(`<saml:SubjectConfirmationData NotOnOrAfter="%s" Recipient="%s"/>`,
		notOnOrAfter, samlTestACSURL)
	if inResponseTo != "" {
		subjectConfirmationData = fmt.Sprintf(`<saml:SubjectConfirmationData InResponseTo="%s" NotOnOrAfter="%s" Recipient="%s"/>`,
			inResponseTo, notOnOrAfter, samlTestACSURL)
	}

	assertionXML := fmt.Sprintf(`<saml:Assertion xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion" xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" ID="%s" IssueInstant="%s" Version="2.0">`+
		`<saml:Issuer>%s</saml:Issuer>`+
		`<saml:Subject><saml:NameID Format="urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress">%s</saml:NameID>`+
		`<saml:SubjectConfirmation Method="urn:oasis:names:tc:SAML:2.0:cm:bearer">%s</saml:SubjectConfirmation></saml:Subject>`+
		`<saml:Conditions NotBefore="%s" NotOnOrAfter="%s">`+
		`<saml:AudienceRestriction><saml:Audience>%s</saml:Audience></saml:AudienceRestriction></saml:Conditions>`+
		`<saml:AuthnStatement AuthnInstant="%s" SessionIndex="_session"><saml:AuthnContext>`+
		`<saml:AuthnContextClassRef>urn:oasis:names:tc:SAML:2.0:ac:classes:PasswordProtectedTransport</saml:AuthnContextClassRef>`+
		`</saml:AuthnContext></saml:AuthnStatement>`+
		`<saml:AttributeStatement><saml:Attribute Name="%s">`+
		`<saml:AttributeValue>%s</saml:AttributeValue></saml:Attribute></saml:AttributeStatement>`+
		`</saml:Assertion>`,
		assertionID, now.Format(time.RFC3339), samlTestIdPIssuer, samlTestSubject, subjectConfirmationData,
		now.Add(-time.Minute).Format(time.RFC3339), notOnOrAfter, samlTestAudience,
		now.Format(time.RFC3339), samlTestEmailAttr, samlTestSubject)

	assertionDoc := etree.NewDocument()
	s.Require().NoError(assertionDoc.ReadFromString(assertionXML))

	signingCtx, err := dsig.NewSigningContext(idp.key, [][]byte{idp.certDER})
	s.Require().NoError(err)
	signedAssertion, err := signingCtx.SignEnveloped(assertionDoc.Root())
	s.Require().NoError(err)

	signedDoc := etree.NewDocument()
	signedDoc.SetRoot(signedAssertion)
	signedAssertionStr, err := signedDoc.WriteToString()
	s.Require().NoError(err)

	responseXML := fmt.Sprintf(`<samlp:Response xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion" ID="%s" IssueInstant="%s" Version="2.0" Destination="%s">`+
		`<saml:Issuer>%s</saml:Issuer>`+
		`<samlp:Status><samlp:StatusCode Value="urn:oasis:names:tc:SAML:2.0:status:Success"/></samlp:Status>`+
		`%s</samlp:Response>`,
		responseID, now.Format(time.RFC3339), samlTestACSURL, samlTestIdPIssuer, signedAssertionStr)

	return base64.StdEncoding.EncodeToString([]byte(responseXML))
}

func (s *samlIdPInitiatedSuite) setUpTenantConfig() config.TenantConfig {
	s.T().Helper()

	cfg := config.DefaultConfig()
	cfg.Saml.Enabled = true
	cfg.Saml.Endpoint = samlTestEndpoint
	cfg.Saml.AudienceUri = samlTestAudience
	cfg.Saml.DefaultRedirectUrl = samlTestRedirectTo
	cfg.Saml.AllowedRedirectURLS = []string{samlTestRedirectTo}
	cfg.Account.AllowSignup = true
	s.Require().NoError(cfg.PostProcess())

	return cfg.TenantConfig
}

func (s *samlIdPInitiatedSuite) setUpHandler(tenantID uuid.UUID, idp *samlTestIdP) *Handler {
	s.T().Helper()

	spCert, err := saml.NewCertificateService(s.Storage).Generate(saml.NewCertificateSubject("hanko-test", ""), saml.CertificateParams{})
	s.Require().NoError(err)
	spCert.TenantID = tenantID
	s.Require().NoError(s.Storage.GetSamlCertificatePersister().Create(spCert))

	_, err = saml.NewSamlProviderManagementService(s.Storage).
		Create(tenantID, "test-idp", idp.metadataURL, "test.example", true, true,
			config.AttributeMap{Email: samlTestEmailAttr})
	s.Require().NoError(err)

	appConfig := config.DefaultConfig().ApplicationConfig

	return NewSamlHandler(appConfig, auditlog.NewLogger(s.Storage, appConfig.AuditLog),
		saml.NewSamlProviderService(s.Storage))
}

// No RelayState, so the callback treats the response as IdP-initiated.
func (s *samlIdPInitiatedSuite) callbackPost(handler *Handler, tenantConfig config.TenantConfig, samlResponse string) string {
	s.T().Helper()

	form := url.Values{}
	form.Set("SAMLResponse", samlResponse)

	req := httptest.NewRequest(http.MethodPost, samlTestACSURL, strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)

	e := echo.New()
	e.Validator = dto.NewCustomValidator()
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("tenant", context.Tenant{ID: uuid.FromStringOrNil(config.DefaultTenantID), Config: tenantConfig})

	s.Require().NoError(handler.CallbackPost(c))
	return rec.Header().Get("Location")
}

func locationHasToken(location string) bool {
	parsed, err := url.Parse(location)
	if err != nil {
		return false
	}
	return parsed.Query().Get(utils.HankoTokenQuery) != ""
}

func (s *samlIdPInitiatedSuite) TestSamlHandler_CallbackPost_IdPInitiated_RejectsReplayUnderANewResponseID() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	s.Require().NoError(s.LoadFixtures("../test/fixtures/thirdparty"))

	tenantID := uuid.FromStringOrNil(config.DefaultTenantID)
	tenantConfig := s.setUpTenantConfig()

	idp, stop := s.setUpIdP()
	defer stop()
	handler := s.setUpHandler(tenantID, idp)

	response := s.samlResponse(idp, "_response_1", "_assertion_1", "")

	location := s.callbackPost(handler, tenantConfig, response)
	s.True(locationHasToken(location), "the first unsolicited response is accepted")

	location = s.callbackPost(handler, tenantConfig, response)
	s.False(locationHasToken(location), "resubmitting the same response is rejected")
	s.Contains(location, "attempting+to+replay+unsolicited+saml+request")

	decoded, err := base64.StdEncoding.DecodeString(response)
	s.Require().NoError(err)
	replayed := strings.Replace(string(decoded), `ID="_response_1"`, `ID="_response_2"`, 1)
	s.Require().NotEqual(string(decoded), replayed)

	location = s.callbackPost(handler, tenantConfig, base64.StdEncoding.EncodeToString([]byte(replayed)))
	s.False(locationHasToken(location), "a new envelope ID does not make the assertion fresh")
	s.Contains(location, "attempting+to+replay+unsolicited+saml+request")
}

func (s *samlIdPInitiatedSuite) TestSamlHandler_CallbackPost_IdPInitiated_RejectsSPInitiatedAssertion() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	s.Require().NoError(s.LoadFixtures("../test/fixtures/thirdparty"))

	tenantID := uuid.FromStringOrNil(config.DefaultTenantID)
	tenantConfig := s.setUpTenantConfig()

	idp, stop := s.setUpIdP()
	defer stop()
	handler := s.setUpHandler(tenantID, idp)

	response := s.samlResponse(idp, "_response_sp", "_assertion_sp", "_authn_request_id")

	location := s.callbackPost(handler, tenantConfig, response)
	s.False(locationHasToken(location), "an SP-initiated assertion is not usable as unsolicited")
	s.Contains(location, "saml+request+is+not+unsolicited")
}

func (s *samlIdPInitiatedSuite) TestSamlHandler_CallbackPost_IdPInitiated_RejectsResponseInResponseTo() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	s.Require().NoError(s.LoadFixtures("../test/fixtures/thirdparty"))

	tenantID := uuid.FromStringOrNil(config.DefaultTenantID)
	tenantConfig := s.setUpTenantConfig()

	idp, stop := s.setUpIdP()
	defer stop()
	handler := s.setUpHandler(tenantID, idp)

	decoded, err := base64.StdEncoding.DecodeString(s.samlResponse(idp, "_response_irt", "_assertion_irt", ""))
	s.Require().NoError(err)
	withInResponseTo := strings.Replace(string(decoded), `ID="_response_irt"`, `ID="_response_irt" InResponseTo="_authn_request_id"`, 1)
	response := base64.StdEncoding.EncodeToString([]byte(withInResponseTo))

	location := s.callbackPost(handler, tenantConfig, response)
	s.False(locationHasToken(location), "an unsolicited Response must not carry InResponseTo")
	s.Contains(location, "saml+request+is+not+unsolicited")
}

func (s *samlIdPInitiatedSuite) TestSamlIDPInitiatedRequestPersister_GetByResponseIDAndIssuer() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	s.Require().NoError(s.LoadFixtures("../test/fixtures/thirdparty"))

	tenantID := uuid.FromStringOrNil(config.DefaultTenantID)
	persister := s.Storage.GetSamlIDPInitiatedRequestPersister()

	request, err := models.NewSamlIDPInitiatedRequest(tenantID, "_assertion_1", samlTestIdPIssuer,
		time.Now().UTC().Add(10*time.Minute))
	s.Require().NoError(err)
	s.Require().NoError(persister.Create(*request))

	found, err := persister.GetByResponseIDAndIssuer("_assertion_1", samlTestIdPIssuer, tenantID)
	s.NoError(err)
	s.Require().NotNil(found)
	s.Equal(request.ID, found.ID)

	found, err = persister.GetByResponseIDAndIssuer("_assertion_2", samlTestIdPIssuer, tenantID)
	s.NoError(err)
	s.Nil(found)
}
