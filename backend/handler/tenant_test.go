package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/test"
)

func TestTenantSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(tenantSuite))
}

type tenantSuite struct {
	test.Suite
}

// tenantTestConfig is test.DefaultConfig with multi-tenancy enabled - the DB-aware SAML
// provider claim mapping check in TenantHandler.Update only runs in that mode (single-tenant
// mode's SAML providers are config-file-based and already covered by ValidateCrossConfig).
func tenantTestConfig() config.Config {
	cfg := test.DefaultConfig
	cfg.ApplicationConfig.MultiTenancy.Enabled = true
	return cfg
}

func (s *tenantSuite) TestUpdate_SamlProviderClaimMapping() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	const tenantID = "00000000-0000-0000-0000-000000000001"

	tests := []struct {
		name               string
		config             string
		expectedStatusCode int
	}{
		{
			name:               "should reject removing a claim still mapped by a saml provider",
			config:             `{"config":{"custom_claims":{"definitions":{}}}}`,
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "should allow keeping the claim a saml provider maps",
			config:             `{"config":{"custom_claims":{"definitions":{"matriculation_number":{"type":"string"}}}}}`,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "should allow adding an unrelated claim alongside the mapped one",
			config:             `{"config":{"custom_claims":{"definitions":{"matriculation_number":{"type":"string"},"is_staff":{"type":"boolean"}}}}}`,
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, currentTest := range tests {
		s.Run(currentTest.name, func() {
			err := s.LoadFixtures("../test/fixtures/tenant_saml_claims")
			s.Require().NoError(err)

			cfg := tenantTestConfig()
			e := NewManagementRouter(&cfg, s.Storage)

			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/tenants/%s", tenantID), strings.NewReader(currentTest.config))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, req)

			s.Require().Equal(currentTest.expectedStatusCode, rec.Code, rec.Body.String())
		})
	}
}
