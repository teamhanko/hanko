package certificate

import (
	"crypto/x509"
	"encoding/json"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
	"github.com/teamhanko/hanko/backend/v3/saml"
	"github.com/teamhanko/hanko/backend/v3/test"
)

func TestCertificateSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(certificateSuite))
}

type certificateSuite struct {
	test.Suite
}

const tenantConfigWithEndpoint = `{"service": {"name": "Test App"}, "saml": {"endpoint_url": "https://auth.example.com"}}`

// withMultiTenancy reads the subject from the tenant's config in the database.
var withMultiTenancy = newSubjectResolver(true, config.TenantConfig{})

// flagDefaults are the parameters the commands get from their flags when none are given on the command line.
var flagDefaults = certificateParams{keyBits: saml.DefaultKeyBits, validityYears: saml.DefaultValidityYears}

func (s *certificateSuite) createTenant(tenantConfig string) uuid.UUID {
	id := uuid.Must(uuid.NewV4())
	now := time.Now()
	s.Require().NoError(s.Storage.GetTenantPersister().Create(models.Tenant{
		ID: id, Config: json.RawMessage(tenantConfig), CreatedAt: now, UpdatedAt: now,
	}))
	return id
}

// storeCertificate stores a certificate for the tenant that has another common name than the tenant's config gives.
func (s *certificateSuite) storeCertificate(tenantID uuid.UUID, params saml.CertificateParams) *models.SamlCertificate {
	cert, err := saml.NewCertificateService(s.Storage).Generate(saml.NewCertificateSubject("old name", ""), params)
	s.Require().NoError(err)
	cert.TenantID = tenantID
	s.Require().NoError(s.Storage.GetSamlCertificatePersister().Create(cert))
	return cert
}

func (s *certificateSuite) storedCertificate(tenantID uuid.UUID) *models.SamlCertificate {
	cert, err := s.Storage.GetSamlCertificatePersister().GetFirst(tenantID)
	s.Require().NoError(err)
	s.Require().NotNil(cert)
	return cert
}

func (s *certificateSuite) parsed(cert *models.SamlCertificate) *x509.Certificate {
	parsed, err := parseCertificate(cert)
	s.Require().NoError(err)
	return parsed
}

func (s *certificateSuite) keyBits(cert *models.SamlCertificate) int {
	bits, err := cert.KeyBits()
	s.Require().NoError(err)
	return bits
}
