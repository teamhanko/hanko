package saml

import (
	"encoding/json"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
	"github.com/teamhanko/hanko/backend/v3/test"
)

const (
	testPostSSOLocation     = "https://idp.example.com/sso/post"
	testRedirectSSOLocation = "https://idp.example.com/sso/redirect"
)

// testMetadataXMLWithSSOServices builds identity provider metadata that lists the given SingleSignOnService elements,
// each given as a binding and a location.
func testMetadataXMLWithSSOServices(ssoServices ...[2]string) string {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata" entityID="https://idp.example.com/metadata">
  <IDPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">`
	for _, ssoService := range ssoServices {
		xml += `
    <SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:` + ssoService[0] + `" Location="` + ssoService[1] + `"/>`
	}
	return xml + `
  </IDPSSODescriptor>
</EntityDescriptor>`
}

func TestSSOURLFromMetadataXML(t *testing.T) {
	post := [2]string{"HTTP-POST", testPostSSOLocation}
	redirect := [2]string{"HTTP-Redirect", testRedirectSSOLocation}

	tests := []struct {
		name     string
		services [][2]string
		expected string
	}{
		{"redirect is preferred over post listed first", [][2]string{post, redirect}, testRedirectSSOLocation},
		{"redirect only", [][2]string{redirect}, testRedirectSSOLocation},
		{"first location if there is no redirect", [][2]string{post}, testPostSSOLocation},
		{"no locations", nil, ""},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expected, ssoURLFromMetadataXML(testMetadataXMLWithSSOServices(testCase.services...)))
		})
	}
}

func TestSamlMetadataServiceSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(samlMetadataServiceSuite))
}

type samlMetadataServiceSuite struct {
	test.Suite
}

// storeTestProviderMetadata creates a provider and stores identity provider metadata for it, with the stored SSO URL
// set independently of the XML (as for a connection stored before the redirect location was preferred). It is shared
// by the suites that need a provider with cached identity provider metadata.
func storeTestProviderMetadata(s *test.Suite, tenantID uuid.UUID, rawXML string, storedSSOURL string) uuid.UUID {
	providerID := uuid.Must(uuid.NewV4())

	err := s.Storage.GetSamlProviderPersister().Create(models.SamlProvider{
		ID:           providerID,
		TenantID:     tenantID,
		Name:         "Test IdP",
		EntityID:     "https://idp.example.com/metadata",
		MetadataURL:  "https://idp.example.com/metadata",
		Domain:       "example.com",
		Enabled:      true,
		AttributeMap: json.RawMessage("{}"),
	})
	s.Require().NoError(err)

	err = s.Storage.GetSamlIDPMetadataPersister().Create(models.SamlIDPMetadata{
		ID:              uuid.Must(uuid.NewV4()),
		TenantID:        tenantID,
		ProviderID:      providerID,
		RawMetadataXML:  rawXML,
		Issuer:          "https://idp.example.com/metadata",
		SSOURL:          storedSSOURL,
		CertificatesPEM: "[]",
	})
	s.Require().NoError(err)

	return providerID
}

func (s *samlMetadataServiceSuite) TestGet_DerivesSSOURLFromStoredXML() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	s.Require().NoError(s.LoadFixtures("../test/fixtures/saml_sync"))
	tenantID := uuid.FromStringOrNil("00000000-0000-0000-0000-000000000001")

	rawXML := testMetadataXMLWithSSOServices(
		[2]string{"HTTP-POST", testPostSSOLocation},
		[2]string{"HTTP-Redirect", testRedirectSSOLocation},
	)
	providerID := storeTestProviderMetadata(&s.Suite, tenantID, rawXML, testPostSSOLocation)

	metadata, err := NewSamlMetadataService(s.Storage).Get(tenantID, providerID)
	s.Require().NoError(err)
	s.Require().NotNil(metadata)

	s.Equal(testRedirectSSOLocation, metadata.SSOURL)
}

func (s *samlMetadataServiceSuite) TestGet_UsesStoredSSOURLIfXMLYieldsNone() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	s.Require().NoError(s.LoadFixtures("../test/fixtures/saml_sync"))
	tenantID := uuid.FromStringOrNil("00000000-0000-0000-0000-000000000001")

	providerID := storeTestProviderMetadata(&s.Suite, tenantID, "", testPostSSOLocation)

	metadata, err := NewSamlMetadataService(s.Storage).Get(tenantID, providerID)
	s.Require().NoError(err)
	s.Require().NotNil(metadata)

	s.Equal(testPostSSOLocation, metadata.SSOURL)
}
