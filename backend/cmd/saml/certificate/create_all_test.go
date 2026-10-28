package certificate

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/saml"
)

func (s *certificateSuite) TestFindTenantsMissingCertificate() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	withoutCertificate := s.createTenant(tenantConfigWithEndpoint)
	withCertificate := s.createTenant(tenantConfigWithEndpoint)
	s.storeCertificate(withCertificate, saml.CertificateParams{})

	missing, err := findTenantsMissingCertificate(s.Storage.GetConnection())

	s.Require().NoError(err)
	s.Contains(missing, withoutCertificate)
	s.NotContains(missing, withCertificate)
	// the default tenant, which the migrations create, has none either
	s.Contains(missing, uuid.Must(uuid.FromString(config.DefaultTenantID)))
}
