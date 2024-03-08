package provider

import (
	"encoding/xml"
	"fmt"
	"github.com/gofrs/uuid"
	saml2 "github.com/russellhaering/gosaml2"
	"github.com/russellhaering/gosaml2/types"
	dsigTypes "github.com/russellhaering/goxmldsig/types"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/config"
	samlConfig "github.com/teamhanko/hanko/backend/ee/saml/config"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/test"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestBaseSamlProviderSuite(t *testing.T) {
	testSuite := new(baseSamlProviderSuite)
	setupMetadataServer(testSuite)

	t.Parallel()
	suite.Run(t, testSuite)
}

type baseSamlProviderSuite struct {
	test.Suite
	server *httptest.Server
}

func setupMetadataServer(s *baseSamlProviderSuite) {
	meta := &types.EntityDescriptor{
		EntityID: "Lorem",
		IDPSSODescriptor: &types.IDPSSODescriptor{
			KeyDescriptors: []types.KeyDescriptor{
				{
					KeyInfo: dsigTypes.KeyInfo{
						X509Data: dsigTypes.X509Data{
							X509Certificates: []dsigTypes.X509Certificate{
								{
									Data: "MIIDCDCCAfCgAwIBAgIBATANBgkqhkiG9w0BAQsFADAnMSUwIwYDVQQDExxIYW5rbyBBdXRoZW50aWNhdGlvbiBTZXJ2aWNlMB4XDTI0MDIxMjA5MjE1OFoXDTI1MDIxMTA5MjE1OFowJzElMCMGA1UEAxMcSGFua28gQXV0aGVudGljYXRpb24gU2VydmljZTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAKFFKmsZSTgOvFiVySkkkNYvuXb7cnv2d74uBPQzCJFU6RY7fRAFOPRmkZ+ICFIOW25/k8S6bn1igcPyAHhMMAuVNO/S1uTAM+A+lkqkyKsjt1L5qrMYbqLXhBd1hMgsEi0jIzGzXFtX4h2B4dd5CtZd0oUHTxWC1Sv7bq0vt5CqcSRaGWN83HHkRySZ+tjtvfTNemzLqvmoQrDBWukL0XJnOs/sbw55sq2oNORQKpwinjGcNoJfrvEgDVXVDrSixHtx5RX03QRn3N7o+dhCDIcp7yHx5n2GEcqLrCY9lniwLZZZ5IgWSOAwZK9N8WLlg6RJad1eIZ8ovdPYSuDMY70CAwEAAaM/MD0wDgYDVR0PAQH/BAQDAgWgMB0GA1UdJQQWMBQGCCsGAQUFBwMBBggrBgEFBQcDAjAMBgNVHRMBAf8EAjAAMA0GCSqGSIb3DQEBCwUAA4IBAQCFU4elOcts4yngWHzolom3t2VC4IQjFwoh59qyXy0cYRgfLrclKdpxgPO556iG/G/UfTbH1sD9cZlmMAIMzFwfn63GaHuQ43QyvHwaq2xLxw1xPM5+kY8QlsourX5RByzJa00P6oLqpa4bHFSWKYoPr8UwHnSrDgC7PNFwxV9RKCmwRrjvEXoCDsRdEuZWHg2Vv4ZlehGR5+NbGlC9uARG2rWtq98YtJkV5z11NOvQYZF38M0IhY16vwQwsVXLeWYHMWDtmiCI2lCVoCPWwwUtu1kBSup/SdIhnhMXrvr5y+bQcMZ/T0zPFTWSTrTwLndIdTowlPsIczvRBs2lFB1j",
								},
							},
						},
					},
				},
			},
			SingleSignOnServices: []types.SingleSignOnService{
				{
					Location: "http://expected.login/",
				},
			},
		},
	}
	data, _ := xml.Marshal(meta)

	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/metadata" {
			s.T().Errorf("Unexpected Request. Expected /metadata")
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}))
}

func (s *baseSamlProviderSuite) TestBaseSamlProvider_Create() {
	// given
	cfg := config.DefaultConfig()
	samlCfg := samlConfig.IdentityProvider{
		Enabled:               true,
		Name:                  "Test Provider",
		Domain:                "hanko.io",
		MetadataUrl:           fmt.Sprintf("%s/metadata", s.server.URL),
		SkipEmailVerification: false,
		AttributeMap: samlConfig.AttributeMap{
			Name: "",
		},
	}

	provider, err := NewBaseSamlProvider(cfg, samlCfg, s.Storage.GetSamlCertificatePersister(), true)
	s.Require().NoError(err)

	s.Assert().NotNil(provider)
	s.Assert().Equal("Lorem", provider.GetService().IdentityProviderIssuer)
	s.Assert().Equal("http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name", provider.GetConfig().AttributeMap.Name)
	s.Assert().Equal("http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname", provider.GetConfig().AttributeMap.GivenName)
	s.Assert().Equal("http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname", provider.GetConfig().AttributeMap.FamilyName)
	s.Assert().Equal("http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress", provider.GetConfig().AttributeMap.Email)
}

func (s *baseSamlProviderSuite) TestBaseSamlProvider_CreateWithoutDefaults() {
	// given
	cfg := config.DefaultConfig()
	samlCfg := samlConfig.IdentityProvider{
		Enabled:               true,
		Name:                  "Test Provider",
		Domain:                "hanko.io",
		MetadataUrl:           fmt.Sprintf("%s/metadata", s.server.URL),
		SkipEmailVerification: false,
		AttributeMap: samlConfig.AttributeMap{
			Name: "",
		},
	}

	provider, err := NewBaseSamlProvider(cfg, samlCfg, s.Storage.GetSamlCertificatePersister(), false)
	s.Require().NoError(err)

	s.Assert().NotNil(provider)
	s.Assert().Equal("Lorem", provider.GetService().IdentityProviderIssuer)
	s.Assert().Equal("", provider.GetConfig().AttributeMap.Name)
	s.Assert().Equal("", provider.GetConfig().AttributeMap.GivenName)
	s.Assert().Equal("", provider.GetConfig().AttributeMap.FamilyName)
	s.Assert().Equal("", provider.GetConfig().AttributeMap.Email)
}

func (s *baseSamlProviderSuite) TestBaseSamlProvider_Fail_CreateOnMetadataError() {
	// given
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/metadata" {
			s.T().Errorf("Unexpected Request. Expected /metadata")
		}

		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte{})
	}))

	cfg := config.DefaultConfig()
	samlCfg := samlConfig.IdentityProvider{
		Enabled:               true,
		Name:                  "Test Provider",
		Domain:                "hanko.io",
		MetadataUrl:           fmt.Sprintf("%s/metadata", server.URL),
		SkipEmailVerification: false,
		AttributeMap: samlConfig.AttributeMap{
			Name: "",
		},
	}

	_, err := NewBaseSamlProvider(cfg, samlCfg, s.Storage.GetSamlCertificatePersister(), false)
	s.Assert().Error(err)
}

func (s *baseSamlProviderSuite) TestBaseSamlProvider_Fail_CreateOnCertificateError() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode")
	}

	// given
	certId, err := uuid.NewV4()
	s.Require().NoError(err)

	now := time.Now()

	err = s.Storage.GetSamlCertificatePersister().Create(&models.SamlCertificate{
		ID:            certId,
		CertData:      "-",
		CertKey:       "--------------------------------",
		EncryptionKey: "--------------------------------",
		CreatedAt:     now,
		UpdatedAt:     now,
	})
	s.Require().NoError(err)

	cfg := config.DefaultConfig()
	samlCfg := samlConfig.IdentityProvider{
		Enabled:               true,
		Name:                  "Test Provider",
		Domain:                "hanko.io",
		MetadataUrl:           fmt.Sprintf("%s/metadata", s.server.URL),
		SkipEmailVerification: false,
		AttributeMap: samlConfig.AttributeMap{
			Name: "",
		},
	}

	_, err = NewBaseSamlProvider(cfg, samlCfg, s.Storage.GetSamlCertificatePersister(), false)
	s.Assert().Error(err)
}

func (s *baseSamlProviderSuite) TestBaseSamlProvider_ProvideMetadataAsXml() {
	// given
	cfg := config.DefaultConfig()
	samlCfg := samlConfig.IdentityProvider{
		Enabled:               true,
		Name:                  "Test Provider",
		Domain:                "hanko.io",
		MetadataUrl:           fmt.Sprintf("%s/metadata", s.server.URL),
		SkipEmailVerification: false,
		AttributeMap: samlConfig.AttributeMap{
			Name: "",
		},
	}

	provider, err := NewBaseSamlProvider(cfg, samlCfg, s.Storage.GetSamlCertificatePersister(), false)
	s.Require().NoError(err)

	metadata, err := provider.ProvideMetadataAsXml()
	s.Require().NoError(err)

	s.Assert().NotEmpty(metadata)
}

func createTestAssertion(expectedValues saml2.Values, expectedExpireDate time.Time, expectedCreationTiome time.Time) saml2.AssertionInfo {
	return saml2.AssertionInfo{
		Values: expectedValues,
		Assertions: []types.Assertion{
			{
				Issuer: &types.Issuer{
					Value: "Lorem",
				},
				Subject: &types.Subject{
					NameID: &types.NameID{
						Value: "Hanko",
					},
				},
				Conditions: &types.Conditions{
					NotOnOrAfter: expectedExpireDate.String(),
				},
			},
		},
		AuthnInstant: &expectedCreationTiome,
	}
}

func (s *baseSamlProviderSuite) TestBaseSamlProvider_GetUserData() {
	// given
	cfg := config.DefaultConfig()
	samlCfg := samlConfig.IdentityProvider{
		Enabled:               true,
		Name:                  "Test Provider",
		Domain:                "hanko.io",
		MetadataUrl:           fmt.Sprintf("%s/metadata", s.server.URL),
		SkipEmailVerification: false,
		AttributeMap: samlConfig.AttributeMap{
			Name:          "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name",
			EmailVerified: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailverified",
		},
	}

	provider, err := NewBaseSamlProvider(cfg, samlCfg, s.Storage.GetSamlCertificatePersister(), false)
	s.Require().NoError(err)

	testValues := make(saml2.Values)
	testValues["http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name"] = types.Attribute{
		Values: []types.AttributeValue{
			{
				Value: "Ipsum",
			},
		},
	}
	testValues["http://schemas.xmlsoap.org/ws/2005/05/identity/claims/email"] = types.Attribute{
		Values: []types.AttributeValue{
			{
				Value: "dev@hanko.io",
			},
		},
	}
	testValues["http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailverified"] = types.Attribute{
		Values: []types.AttributeValue{
			{
				Value: "true",
			},
		},
	}

	testCreateTime := time.Now()
	testExpireTime := testCreateTime.Add(time.Hour * 24 * 7)
	testAssertion := createTestAssertion(testValues, testCreateTime, testExpireTime)

	// when
	userdata := provider.GetUserData(&testAssertion)

	// then
	s.Assert().Equal(testAssertion.Assertions[0].Issuer.Value, userdata.Metadata.Issuer)
	s.Assert().Equal(testAssertion.Assertions[0].Subject.NameID.Value, userdata.Metadata.Subject)
	s.Assert().Equal(testAssertion.Values.Get("http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name"), userdata.Metadata.Name)
	s.Assert().Equal(testAssertion.Values.Get("http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"), userdata.Metadata.Email)
	s.Assert().True(userdata.Metadata.EmailVerified)
	s.Assert().True(userdata.Emails[0].Verified)
}

func (s *baseSamlProviderSuite) TestBaseSamlProvider_GetUserData_WithoutVerification() {
	// given
	cfg := config.DefaultConfig()
	samlCfg := samlConfig.IdentityProvider{
		Enabled:               true,
		Name:                  "Test Provider",
		Domain:                "hanko.io",
		MetadataUrl:           fmt.Sprintf("%s/metadata", s.server.URL),
		SkipEmailVerification: false,
		AttributeMap: samlConfig.AttributeMap{
			Name:          "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name",
			EmailVerified: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailverified",
		},
	}

	provider, err := NewBaseSamlProvider(cfg, samlCfg, s.Storage.GetSamlCertificatePersister(), false)
	s.Require().NoError(err)

	testValues := make(saml2.Values)
	testValues["http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name"] = types.Attribute{
		Values: []types.AttributeValue{
			{
				Value: "Ipsum",
			},
		},
	}
	testValues["http://schemas.xmlsoap.org/ws/2005/05/identity/claims/email"] = types.Attribute{
		Values: []types.AttributeValue{
			{
				Value: "dev@hanko.io",
			},
		},
	}

	testCreateTime := time.Now()
	testExpireTime := testCreateTime.Add(time.Hour * 24 * 7)
	testAssertion := createTestAssertion(testValues, testCreateTime, testExpireTime)

	// when
	userdata := provider.GetUserData(&testAssertion)

	// then
	s.Assert().False(userdata.Metadata.EmailVerified)
	s.Assert().False(userdata.Emails[0].Verified)
}

func (s *baseSamlProviderSuite) TestBaseSamlProvider_GetUserData_WithSkipVerification() {
	// given
	cfg := config.DefaultConfig()
	samlCfg := samlConfig.IdentityProvider{
		Enabled:               true,
		Name:                  "Test Provider",
		Domain:                "hanko.io",
		MetadataUrl:           fmt.Sprintf("%s/metadata", s.server.URL),
		SkipEmailVerification: true,
		AttributeMap: samlConfig.AttributeMap{
			Name:          "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name",
			EmailVerified: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailverified",
		},
	}

	provider, err := NewBaseSamlProvider(cfg, samlCfg, s.Storage.GetSamlCertificatePersister(), false)
	s.Require().NoError(err)

	testValues := make(saml2.Values)
	testValues["http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name"] = types.Attribute{
		Values: []types.AttributeValue{
			{
				Value: "Ipsum",
			},
		},
	}
	testValues["http://schemas.xmlsoap.org/ws/2005/05/identity/claims/email"] = types.Attribute{
		Values: []types.AttributeValue{
			{
				Value: "dev@hanko.io",
			},
		},
	}

	testCreateTime := time.Now()
	testExpireTime := testCreateTime.Add(time.Hour * 24 * 7)
	testAssertion := createTestAssertion(testValues, testCreateTime, testExpireTime)

	// when
	userdata := provider.GetUserData(&testAssertion)

	// then
	s.Assert().True(userdata.Metadata.EmailVerified)
	s.Assert().False(userdata.Emails[0].Verified)
}

func (s *baseSamlProviderSuite) TestBaseSamlProvider_GetDomain() {
	// given
	cfg := config.DefaultConfig()
	samlCfg := samlConfig.IdentityProvider{
		Enabled:               true,
		Name:                  "Test Provider",
		Domain:                "hanko.io",
		MetadataUrl:           fmt.Sprintf("%s/metadata", s.server.URL),
		SkipEmailVerification: true,
		AttributeMap: samlConfig.AttributeMap{
			Name:          "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name",
			EmailVerified: "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailverified",
		},
	}

	provider, err := NewBaseSamlProvider(cfg, samlCfg, s.Storage.GetSamlCertificatePersister(), false)
	s.Require().NoError(err)

	s.Assert().Equal(samlCfg.Domain, provider.GetDomain())
}
