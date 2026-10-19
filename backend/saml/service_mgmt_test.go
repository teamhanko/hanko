package saml

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	samlConfig "github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/test"
)

// testMetadataXML is a minimal, validly-shaped SAML IdP metadata document - just enough for
// SamlMetadataService.FetchAndParse (which only reads EntityID and the SSO service list) to
// succeed. No signing certificate, since CreateFromConfig never inspects one.
const testMetadataXML = `<?xml version="1.0" encoding="UTF-8"?>
<EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata" entityID="https://idp.example.com/metadata">
  <IDPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
    <SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect" Location="https://idp.example.com/sso"/>
  </IDPSSODescriptor>
</EntityDescriptor>`

func TestSamlProviderManagementServiceSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(samlProviderManagementServiceSuite))
}

type samlProviderManagementServiceSuite struct {
	test.Suite
}

// CreateFromConfig is single-tenant mode's startup sync path (sync.go's
// SyncProviderConfigToDatabase calls it once per configured identity provider) - it upserts
// config-file-defined SAML identity providers, including their custom-claim
// attribute_map.custom, into the same saml_providers DB table runtime login reads from in both
// tenancy modes. Untested until now: single-tenant mode never goes through the DB-backed CRUD
// endpoints (backend/handler/saml_provider.go) that already had coverage.
func (s *samlProviderManagementServiceSuite) TestCreateFromConfig_PersistsAndUpdatesAttributeMapCustom() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	metadataServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(testMetadataXML))
	}))
	defer metadataServer.Close()

	err := s.LoadFixtures("../test/fixtures/saml_sync")
	s.Require().NoError(err)

	tenantID := uuid.FromStringOrNil("00000000-0000-0000-0000-000000000001")
	service := NewSamlProviderManagementService(s.Storage)

	idpConfig := samlConfig.IdentityProvider{
		Enabled:     true,
		Name:        "Uni A",
		Domain:      "uni-a.example.com",
		MetadataUrl: metadataServer.URL,
		AttributeMap: samlConfig.AttributeMap{
			Custom: map[string]string{"matriculation_number": "mat_nr"},
		},
	}

	err = service.CreateFromConfig(tenantID, idpConfig)
	s.Require().NoError(err)

	created, err := s.Storage.GetSamlProviderPersister().GetByDomain(tenantID, "uni-a.example.com")
	s.Require().NoError(err)
	s.Require().NotNil(created)

	var attrMap samlConfig.AttributeMap
	s.Require().NoError(json.Unmarshal(created.AttributeMap, &attrMap))
	s.Equal(map[string]string{"matriculation_number": "mat_nr"}, attrMap.Custom)

	// Re-running with a changed mapping (simulating a config-file edit + restart) must hit the
	// existing-provider update branch, not create a duplicate row for the same domain.
	idpConfig.AttributeMap.Custom = map[string]string{"matriculation_number": "student_nr", "affiliation": "aff"}
	err = service.CreateFromConfig(tenantID, idpConfig)
	s.Require().NoError(err)

	updated, err := s.Storage.GetSamlProviderPersister().GetByDomain(tenantID, "uni-a.example.com")
	s.Require().NoError(err)
	s.Require().NotNil(updated)
	s.Equal(created.ID, updated.ID, "must update the existing provider, not create a second one")

	var updatedAttrMap samlConfig.AttributeMap
	s.Require().NoError(json.Unmarshal(updated.AttributeMap, &updatedAttrMap))
	s.Equal(map[string]string{"matriculation_number": "student_nr", "affiliation": "aff"}, updatedAttrMap.Custom)
}
