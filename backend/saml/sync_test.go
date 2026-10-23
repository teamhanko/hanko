package saml

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/persistence"
	"github.com/teamhanko/hanko/backend/v3/test"
)

func TestSamlSyncSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(samlSyncSuite))
}

type samlSyncSuite struct {
	test.Suite
}

// failingTransactions makes every transaction fail, which is what creating the certificate runs in.
type failingTransactions struct {
	persistence.Persister
}

func (failingTransactions) Transaction(func(tx *pop.Connection) error) error {
	return errors.New("transaction failed")
}

func syncConfig(metadataUrl string) config.Config {
	cfg := test.DefaultConfig
	cfg.MultiTenancy.Enabled = false
	cfg.Service.Name = "Test App"
	cfg.Saml.Enabled = true
	cfg.Saml.Endpoint = "https://auth.example.com"
	cfg.Saml.IdentityProviders = []config.IdentityProvider{
		{Enabled: true, Name: "Uni A", Domain: "uni-a.example.com", MetadataUrl: metadataUrl},
	}
	return cfg
}

func (s *samlSyncSuite) TestSyncProviderConfigToDatabase_CreatesCertificateFromConfig() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	metadataServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(testMetadataXML))
	}))
	defer metadataServer.Close()

	tenantID := uuid.FromStringOrNil(config.DefaultTenantID)
	existing, err := s.Storage.GetSamlCertificatePersister().GetFirst(tenantID)
	s.Require().NoError(err)
	s.Require().Nil(existing)

	cfg := syncConfig(metadataServer.URL)
	s.Require().NoError(SyncProviderConfigToDatabase(&cfg, s.Storage))

	cert, err := s.Storage.GetSamlCertificatePersister().GetFirst(tenantID)
	s.Require().NoError(err)
	s.Require().NotNil(cert)
	parsed := parseCert(s.T(), cert.CertData)
	s.Equal("auth.example.com", parsed.Subject.CommonName)
	s.Equal([]string{"auth.example.com"}, parsed.DNSNames)
}

func (s *samlSyncSuite) TestSyncProviderConfigToDatabase_ReturnsCertificateError() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	// the metadata is never fetched, because the sync stops at the certificate
	cfg := syncConfig("https://idp.example.com/metadata")

	err := SyncProviderConfigToDatabase(&cfg, failingTransactions{s.Storage})

	s.Require().Error(err)
	s.Contains(err.Error(), "failed to ensure SAML certificate exists")
}
