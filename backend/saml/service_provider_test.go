package saml

import (
	"compress/flate"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"io"
	"net/url"
	"strings"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
	"github.com/teamhanko/hanko/backend/v3/test"
)

const testRSASHA256SigAlg = "http://www.w3.org/2001/04/xmldsig-more#rsa-sha256"

func TestSamlProviderServiceSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(samlProviderServiceSuite))
}

type samlProviderServiceSuite struct {
	test.Suite
}

// getAuthUrl creates a provider with the metadata of an identity provider that lists different post and redirect
// locations, and a service provider certificate, and returns the url to start the login at that provider together with
// the service provider certificate.
func (s *samlProviderServiceSuite) getAuthUrl(signAuthnRequests bool) (*url.URL, *x509.Certificate) {
	s.Require().NoError(s.LoadFixtures("../test/fixtures/saml_sync"))
	tenantID := uuid.FromStringOrNil("00000000-0000-0000-0000-000000000001")

	// The stored SSO URL is the post location, as for a connection stored before the redirect location was preferred
	rawXML := testMetadataXMLWithSSOServices(
		[2]string{"HTTP-POST", testPostSSOLocation},
		[2]string{"HTTP-Redirect", testRedirectSSOLocation},
	)
	providerID := storeTestProviderMetadata(&s.Suite, tenantID, rawXML, testPostSSOLocation)

	certificate, err := models.NewSamlCertificate("Test Service")
	s.Require().NoError(err)
	certificate.TenantID = tenantID
	s.Require().NoError(s.Storage.GetSamlCertificatePersister().Create(certificate))

	cfg := config.Config{
		ApplicationConfig: config.ApplicationConfig{
			SecretKeys: []string{"thirty-two-byte-long-test-secret"},
		},
		TenantConfig: config.TenantConfig{
			Saml: config.Saml{
				Endpoint:           "https://auth.example.com",
				DefaultRedirectUrl: "https://app.example.com",
				Options:            config.Options{SignAuthnRequests: signAuthnRequests},
			},
		},
	}
	s.Require().NoError(cfg.TenantConfig.Saml.PostProcess())

	authUrl, err := NewSamlProviderService(s.Storage).GetAuthUrl(tenantID, cfg, providerID, "https://app.example.com", false)
	s.Require().NoError(err)

	parsedUrl, err := url.Parse(authUrl)
	s.Require().NoError(err)

	block, _ := pem.Decode([]byte(certificate.CertData))
	s.Require().NotNil(block)
	parsedCertificate, err := x509.ParseCertificate(block.Bytes)
	s.Require().NoError(err)

	return parsedUrl, parsedCertificate
}

// inflateAuthnRequest decodes the SAMLRequest query parameter of a redirect binding url
func (s *samlProviderServiceSuite) inflateAuthnRequest(authUrl *url.URL) string {
	deflated, err := base64.StdEncoding.DecodeString(authUrl.Query().Get("SAMLRequest"))
	s.Require().NoError(err)

	inflated, err := io.ReadAll(flate.NewReader(strings.NewReader(string(deflated))))
	s.Require().NoError(err)

	return string(inflated)
}

func (s *samlProviderServiceSuite) TestGetAuthUrl_SignsTheRedirectUrl() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	authUrl, certificate := s.getAuthUrl(true)

	// The request is sent to the redirect location, which is derived from the stored XML and not the stored SSO URL
	s.Equal(testRedirectSSOLocation, authUrl.Scheme+"://"+authUrl.Host+authUrl.Path)

	// The signature is in the query, not in the XML
	authnRequest := s.inflateAuthnRequest(authUrl)
	s.NotContains(authnRequest, "Signature")
	s.Contains(authnRequest, `Destination="`+testRedirectSSOLocation+`"`)

	s.Equal(testRSASHA256SigAlg, authUrl.Query().Get("SigAlg"))

	// The signature is over the query parameters exactly as sent, in the order SAMLRequest, RelayState, SigAlg
	rawParams := map[string]string{}
	for _, rawParam := range strings.Split(authUrl.RawQuery, "&") {
		name, value, _ := strings.Cut(rawParam, "=")
		rawParams[name] = value
	}
	signedString := "SAMLRequest=" + rawParams["SAMLRequest"] + "&RelayState=" + rawParams["RelayState"] + "&SigAlg=" + rawParams["SigAlg"]

	signature, err := base64.StdEncoding.DecodeString(authUrl.Query().Get("Signature"))
	s.Require().NoError(err)

	digest := sha256.Sum256([]byte(signedString))
	publicKey, ok := certificate.PublicKey.(*rsa.PublicKey)
	s.Require().True(ok)
	s.NoError(rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, digest[:], signature))
}

func (s *samlProviderServiceSuite) TestGetAuthUrl_DoesNotSignTheRedirectUrlIfSigningIsDisabled() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	authUrl, _ := s.getAuthUrl(false)

	s.Equal(testRedirectSSOLocation, authUrl.Scheme+"://"+authUrl.Host+authUrl.Path)
	s.NotContains(s.inflateAuthnRequest(authUrl), "Signature")
	s.False(authUrl.Query().Has("SigAlg"))
	s.False(authUrl.Query().Has("Signature"))
}
