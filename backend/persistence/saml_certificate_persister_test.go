package persistence_test

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
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

func parseCertificate(s *samlCertificatePersisterSuite, certPEM string) *x509.Certificate {
	block, _ := pem.Decode([]byte(certPEM))
	s.Require().NotNil(block)
	cert, err := x509.ParseCertificate(block.Bytes)
	s.Require().NoError(err)
	return cert
}

func (s *samlCertificatePersisterSuite) TestRenew_PersistsNewCertificateWithSameKey() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	tenantID := uuid.Must(uuid.FromString(defaultTenantID))

	cert, err := models.NewSamlCertificate("test")
	s.Require().NoError(err)
	cert.TenantID = tenantID

	// replace the certificate with one that expired a year ago, signed by the same key
	keyPEM, err := cert.DecryptCertKey()
	s.Require().NoError(err)
	keyBlock, _ := pem.Decode(keyPEM)
	s.Require().NotNil(keyBlock)
	privateKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	s.Require().NoError(err)
	expiredCert, err := models.GenerateCertificate("test", privateKey, time.Now().AddDate(-11, 0, 0))
	s.Require().NoError(err)
	cert.CertData = expiredCert

	persister := s.Storage.GetSamlCertificatePersister()
	s.Require().NoError(persister.Create(cert))
	encryptedKey := cert.CertKey

	s.Require().NoError(persister.Renew(cert, "test"))

	stored, err := persister.GetFirst(tenantID)
	s.Require().NoError(err)
	s.Require().NotNil(stored)

	s.NotEqual(expiredCert, stored.CertData, "renewed certificate must be persisted")
	s.True(parseCertificate(s, stored.CertData).NotAfter.After(time.Now()))

	// the key pair is kept, only the certificate changes
	s.Equal(encryptedKey, stored.CertKey)
	oldPublicKey := parseCertificate(s, expiredCert).PublicKey.(*rsa.PublicKey)
	s.True(oldPublicKey.Equal(parseCertificate(s, stored.CertData).PublicKey))
}
