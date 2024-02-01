package provider

import (
	"encoding/xml"
	"fmt"
	"github.com/russellhaering/gosaml2/types"
	dsigTypes "github.com/russellhaering/goxmldsig/types"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/config"
	samlConfig "github.com/teamhanko/hanko/backend/ee/saml/config"
	"github.com/teamhanko/hanko/backend/test"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProviderSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(providerSuite))
}

type providerSuite struct {
	test.Suite
}

func (s *providerSuite) TestProvider_loadCertificate() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode")
	}

	err := s.LoadFixtures("../../../test/fixtures/saml")
	s.Require().NoError(err)

	cfg := config.DefaultConfig()

	store, err := loadCertificate(cfg, s.Storage.GetSamlCertificatePersister())
	s.Require().NoError(err)

	s.Assert().NotNil(store)
}

func (s *providerSuite) TestProvider_Create_Cert_On_Load() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode")
	}

	cfg := config.DefaultConfig()

	store, err := loadCertificate(cfg, s.Storage.GetSamlCertificatePersister())
	s.Require().NoError(err)

	s.Assert().NotNil(store)

	cert, err := s.Storage.GetSamlCertificatePersister().GetFirst()
	s.Require().NoError(err)

	s.Assert().NotNil(cert)
	s.Assert().NotNil(cert.CertData)
}

func (s *providerSuite) TestProvider_Fetch_IDP_Metadata() {
	// given
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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/metadata" {
			s.T().Errorf("Unexpected Request. Expected /metadata")
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}))

	cfg := samlConfig.IdentityProvider{
		MetadataUrl: fmt.Sprintf("%s/metadata", server.URL),
	}

	// when
	fetchedMeta, err := fetchIdpMetadata(cfg)
	s.Require().NoError(err)

	s.Assert().Equal(meta.IDPSSODescriptor.SingleSignOnServices[0].Location, fetchedMeta.SingleSignOnUrl)
	s.Assert().Equal(meta.EntityID, fetchedMeta.Issuer)
	s.Assert().Equal(1, len(fetchedMeta.certs.Roots))
}

func (s *providerSuite) TestProvider_FAIL_Fetch_IDP_Metadata_With_Non_Xml() {
	// given
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/metadata" {
			s.T().Errorf("Unexpected Request. Expected /metadata")
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{ \"iam\": \"JSON\"}"))
	}))

	cfg := samlConfig.IdentityProvider{
		MetadataUrl: fmt.Sprintf("%s/metadata", server.URL),
	}

	// when
	_, err := fetchIdpMetadata(cfg)
	s.Assert().ErrorContains(err, "unable to unmarshal idp metadata response")
}

func (s *providerSuite) TestProvider_FAIL_Fetch_IDP_Metadata_With_Wrong_Status_Code() {
	// given
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/metadata" {
			s.T().Errorf("Unexpected Request. Expected /metadata")
		}

		w.WriteHeader(http.StatusNoContent)
		_, _ = w.Write(nil)
	}))

	cfg := samlConfig.IdentityProvider{
		MetadataUrl: fmt.Sprintf("%s/metadata", server.URL),
	}

	// when
	_, err := fetchIdpMetadata(cfg)
	s.Assert().ErrorContains(err, "request for idp metadata failed with status code")
}

func (s *providerSuite) TestProvider_Fail_Fetch_IDP_Metadata_No_Parsable_Cert() {
	// given
	meta := &types.EntityDescriptor{
		EntityID: "Lorem",
		IDPSSODescriptor: &types.IDPSSODescriptor{
			KeyDescriptors: []types.KeyDescriptor{
				{
					KeyInfo: dsigTypes.KeyInfo{
						X509Data: dsigTypes.X509Data{
							X509Certificates: []dsigTypes.X509Certificate{
								{
									Data: "DATA",
								},
							},
						},
					},
				},
			},
		},
	}
	data, _ := xml.Marshal(meta)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/metadata" {
			s.T().Errorf("Unexpected Request. Expected /metadata")
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}))

	cfg := samlConfig.IdentityProvider{
		MetadataUrl: fmt.Sprintf("%s/metadata", server.URL),
	}

	// when
	_, err := fetchIdpMetadata(cfg)
	s.Assert().Error(err)

	s.Assert().ErrorContains(err, "malformed certificate")
}

func (s *providerSuite) TestProvider_Parse_Certificate() {
	// given
	cert := dsigTypes.X509Certificate{
		Data: "MIIDCDCCAfCgAwIBAgIBATANBgkqhkiG9w0BAQsFADAnMSUwIwYDVQQDExxIYW5rbyBBdXRoZW50aWNhdGlvbiBTZXJ2aWNlMB4XDTI0MDIxMjA5MjE1OFoXDTI1MDIxMTA5MjE1OFowJzElMCMGA1UEAxMcSGFua28gQXV0aGVudGljYXRpb24gU2VydmljZTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAKFFKmsZSTgOvFiVySkkkNYvuXb7cnv2d74uBPQzCJFU6RY7fRAFOPRmkZ+ICFIOW25/k8S6bn1igcPyAHhMMAuVNO/S1uTAM+A+lkqkyKsjt1L5qrMYbqLXhBd1hMgsEi0jIzGzXFtX4h2B4dd5CtZd0oUHTxWC1Sv7bq0vt5CqcSRaGWN83HHkRySZ+tjtvfTNemzLqvmoQrDBWukL0XJnOs/sbw55sq2oNORQKpwinjGcNoJfrvEgDVXVDrSixHtx5RX03QRn3N7o+dhCDIcp7yHx5n2GEcqLrCY9lniwLZZZ5IgWSOAwZK9N8WLlg6RJad1eIZ8ovdPYSuDMY70CAwEAAaM/MD0wDgYDVR0PAQH/BAQDAgWgMB0GA1UdJQQWMBQGCCsGAQUFBwMBBggrBgEFBQcDAjAMBgNVHRMBAf8EAjAAMA0GCSqGSIb3DQEBCwUAA4IBAQCFU4elOcts4yngWHzolom3t2VC4IQjFwoh59qyXy0cYRgfLrclKdpxgPO556iG/G/UfTbH1sD9cZlmMAIMzFwfn63GaHuQ43QyvHwaq2xLxw1xPM5+kY8QlsourX5RByzJa00P6oLqpa4bHFSWKYoPr8UwHnSrDgC7PNFwxV9RKCmwRrjvEXoCDsRdEuZWHg2Vv4ZlehGR5+NbGlC9uARG2rWtq98YtJkV5z11NOvQYZF38M0IhY16vwQwsVXLeWYHMWDtmiCI2lCVoCPWwwUtu1kBSup/SdIhnhMXrvr5y+bQcMZ/T0zPFTWSTrTwLndIdTowlPsIczvRBs2lFB1j",
	}

	// when
	parsedCert, err := parseCertificate(0, cert)
	s.Require().NoError(err)

	s.Assert().NotNil(parsedCert)
	s.Equal("Hanko Authentication Service", parsedCert.Issuer.CommonName)
}

func (s *providerSuite) TestProvider_Fail_Parse_Empty_Certificate() {
	// given
	cert := dsigTypes.X509Certificate{
		Data: "",
	}

	// when
	_, err := parseCertificate(0, cert)
	s.Assert().Error(err)
	s.Assert().ErrorContains(err, "metadata contains an empty certificate at index")
}

func (s *providerSuite) TestProvider_Fail_Parse_NoB64_Certificate() {
	// given
	cert := dsigTypes.X509Certificate{
		Data: "----",
	}

	// when
	_, err := parseCertificate(0, cert)
	s.Assert().Error(err)
	s.Assert().ErrorContains(err, "unable to decode certificate at index")
}

func (s *providerSuite) TestProvider_Fail_Parse_Wrong_Certificate() {
	// given
	cert := dsigTypes.X509Certificate{
		Data: "DIIDCDCCAfCgAwIBAgIBATANBgkqhkiG9w0BAQsFADAnMSUwIwYDVQQDExxIYW5rbyBBdXRoZW50aWNhdGlvbiBTZXJ2aWNlMB4XDTI0MDIxMjA5MjE1OFoXDTI1MDIxMTA5MjE1OFowJzElMCMGA1UEAxMcSGFua28gQXV0aGVudGljYXRpb24gU2VydmljZTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAKFFKmsZSTgOvFiVySkkkNYvuXb7cnv2d74uBPQzCJFU6RY7fRAFOPRmkZ+ICFIOW25/k8S6bn1igcPyAHhMMAuVNO/S1uTAM+A+lkqkyKsjt1L5qrMYbqLXhBd1hMgsEi0jIzGzXFtX4h2B4dd5CtZd0oUHTxWC1Sv7bq0vt5CqcSRaGWN83HHkRySZ+tjtvfTNemzLqvmoQrDBWukL0XJnOs/sbw55sq2oNORQKpwinjGcNoJfrvEgDVXVDrSixHtx5RX03QRn3N7o+dhCDIcp7yHx5n2GEcqLrCY9lniwLZZZ5IgWSOAwZK9N8WLlg6RJad1eIZ8ovdPYSuDMY70CAwEAAaM/MD0wDgYDVR0PAQH/BAQDAgWgMB0GA1UdJQQWMBQGCCsGAQUFBwMBBggrBgEFBQcDAjAMBgNVHRMBAf8EAjAAMA0GCSqGSIb3DQEBCwUAA4IBAQCFU4elOcts4yngWHzolom3t2VC4IQjFwoh59qyXy0cYRgfLrclKdpxgPO556iG/G/UfTbH1sD9cZlmMAIMzFwfn63GaHuQ43QyvHwaq2xLxw1xPM5+kY8QlsourX5RByzJa00P6oLqpa4bHFSWKYoPr8UwHnSrDgC7PNFwxV9RKCmwRrjvEXoCDsRdEuZWHg2Vv4ZlehGR5+NbGlC9uARG2rWtq98YtJkV5z11NOvQYZF38M0IhY16vwQwsVXLeWYHMWDtmiCI2lCVoCPWwwUtu1kBSup/SdIhnhMXrvr5y+bQcMZ/T0zPFTWSTrTwLndIdTowlPsIczvRBs2lFB1j",
	}

	// when
	_, err := parseCertificate(0, cert)
	s.Assert().Error(err)
	s.Assert().ErrorContains(err, "unable to parse certificate at index")
}

func (s *providerSuite) TestProvider_GetProvider() {
	// given
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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/metadata" {
			s.T().Errorf("Unexpected Request. Expected /metadata")
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}))

	providerName := "lorem"
	cfg := config.DefaultConfig()
	samlCfg := samlConfig.IdentityProvider{
		Enabled:               true,
		Name:                  "Test Provider",
		Domain:                "hanko.io",
		MetadataUrl:           fmt.Sprintf("%s/metadata", server.URL),
		SkipEmailVerification: false,
		AttributeMap:          samlConfig.AttributeMap{},
	}

	// when
	provider, err := GetProvider(providerName, cfg, samlCfg, s.Storage.GetSamlCertificatePersister())
	s.Require().NoError(err)

	s.Assert().Equal("http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name", provider.GetConfig().AttributeMap.Name)
	s.Assert().Equal("http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname", provider.GetConfig().AttributeMap.GivenName)
	s.Assert().Equal("http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname", provider.GetConfig().AttributeMap.FamilyName)
	s.Assert().Equal("http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress", provider.GetConfig().AttributeMap.Email)
}

func (s *providerSuite) TestProvider_GetProvider_Auth0() {
	// given
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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/metadata" {
			s.T().Errorf("Unexpected Request. Expected /metadata")
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}))

	providerName := "auth0"
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

	// when
	provider, err := GetProvider(providerName, cfg, samlCfg, s.Storage.GetSamlCertificatePersister())
	s.Require().NoError(err)

	s.Assert().Equal("http://schemas.auth0.com/name", provider.GetConfig().AttributeMap.Name)
	s.Assert().Equal("http://schemas.auth0.com/email", provider.GetConfig().AttributeMap.Email)
	s.Assert().Equal("http://schemas.auth0.com/email_verified", provider.GetConfig().AttributeMap.EmailVerified)
	s.Assert().Equal("http://schemas.auth0.com/nickname", provider.GetConfig().AttributeMap.NickName)
	s.Assert().Equal("http://schemas.auth0.com/picture", provider.GetConfig().AttributeMap.Picture)
	s.Assert().Equal("http://schemas.auth0.com/updated_at", provider.GetConfig().AttributeMap.UpdatedAt)
}
