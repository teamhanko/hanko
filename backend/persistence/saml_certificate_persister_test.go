package persistence_test

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v3/saml"
	"github.com/teamhanko/hanko/backend/v3/test"
)

func TestSamlCertificatePersisterSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(samlCertificatePersisterSuite))
}

type samlCertificatePersisterSuite struct {
	test.Suite
}

// defaultTenantID is the tenant inserted by the add_tenant_id migration.
const defaultTenantID = "00000000-0000-0000-0000-000000000001"

func (s *samlCertificatePersisterSuite) TestUpdate_PersistsCertificateAndKey() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	tenantID := uuid.Must(uuid.FromString(defaultTenantID))
	service := saml.NewCertificateService(s.Storage)

	cert, err := service.Generate("test", saml.CertificateParams{})
	s.Require().NoError(err)
	cert.TenantID = tenantID
	persister := s.Storage.GetSamlCertificatePersister()
	s.Require().NoError(persister.Create(cert))

	replacement, err := service.Generate("test", saml.CertificateParams{KeyBits: 3072})
	s.Require().NoError(err)
	key, err := replacement.PrivateKey()
	s.Require().NoError(err)

	cert.CertData = replacement.CertData
	s.Require().NoError(cert.SetPrivateKey(key))
	s.Require().NoError(persister.Update(cert))

	stored, err := persister.GetFirst(tenantID)
	s.Require().NoError(err)
	s.Require().NotNil(stored)
	s.Equal(replacement.CertData, stored.CertData, "updated certificate must be persisted")
	bits, err := stored.KeyBits()
	s.Require().NoError(err)
	s.Equal(3072, bits, "updated key must be persisted")
}
