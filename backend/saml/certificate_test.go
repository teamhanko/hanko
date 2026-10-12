package saml

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
	"github.com/teamhanko/hanko/backend/v3/test"
)

func TestCertificateExpirySuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(certificateExpirySuite))
}

type certificateExpirySuite struct {
	test.Suite
}

// createCertificate stores a certificate for a new tenant that ends at notAfter.
func (s *certificateExpirySuite) createCertificate(key *rsa.PrivateKey, notAfter time.Time) uuid.UUID {
	tenantID := uuid.Must(uuid.NewV4())
	s.Require().NoError(s.Storage.GetTenantPersister().Create(models.Tenant{ID: tenantID, Config: json.RawMessage("{}")}))

	// generated certificates are valid for 10 years
	certData, err := models.GenerateCertificate("test", key, notAfter.Add(-3650*24*time.Hour))
	s.Require().NoError(err)

	s.Require().NoError(s.Storage.GetSamlCertificatePersister().Create(&models.SamlCertificate{
		ID:            uuid.Must(uuid.NewV4()),
		TenantID:      tenantID,
		CertData:      certData,
		CertKey:       "unused",
		EncryptionKey: "unused-unused-unused-unused-unused",
	}))

	return tenantID
}

func (s *certificateExpirySuite) TestFindExpiringCertificates() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	s.Require().NoError(err)

	now := time.Now()
	expiringSoon := s.createCertificate(key, now.AddDate(0, 0, 30))
	expired := s.createCertificate(key, now.AddDate(0, 0, -30))
	s.createCertificate(key, now.AddDate(0, 0, 120))

	got, err := findExpiringCertificates(s.Storage.GetConnection(), now, 60*24*time.Hour)
	s.Require().NoError(err)

	s.Require().Len(got, 2)
	s.Equal(expired, got[0].TenantID, "soonest first")
	s.Equal(expiringSoon, got[1].TenantID)
}
