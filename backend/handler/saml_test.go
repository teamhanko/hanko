package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
	"github.com/teamhanko/hanko/backend/v3/test"
)

func TestSamlSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(samlSuite))
}

type samlSuite struct {
	test.Suite
}

// defaultTenantID is the tenant inserted by the add_tenant_id migration.
const defaultTenantID = "00000000-0000-0000-0000-000000000001"

func (s *samlSuite) createCertificate() {
	cert, err := models.NewSamlCertificate("test")
	s.Require().NoError(err)
	cert.TenantID = uuid.Must(uuid.FromString(defaultTenantID))
	s.Require().NoError(s.Storage.GetSamlCertificatePersister().Create(cert))
}

func (s *samlSuite) get(path string) *httptest.ResponseRecorder {
	cfg := test.DefaultConfig
	cfg.TenantConfig.Saml.Enabled = true
	cfg.TenantConfig.Saml.Endpoint = "https://hanko.example.com"
	e := NewPublicRouter(&cfg, s.Storage, nil, nil)

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

// The metadata is tenant-wide, so it must be available without any SAML provider
// configured and without a domain.
func (s *samlSuite) TestMetadata_WithoutProviderOrDomain() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	s.createCertificate()

	rec := s.get("/saml/metadata")

	s.Require().Equal(http.StatusOK, rec.Code, rec.Body.String())
	s.Contains(rec.Body.String(), `entityID="https://hanko.example.com"`)
	s.Contains(rec.Body.String(), "https://hanko.example.com/saml/callback")
	s.Contains(rec.Body.String(), `use="signing"`)
}

// Metadata URLs handed out before the domain was dropped must keep working.
func (s *samlSuite) TestMetadata_IgnoresDomain() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	s.createCertificate()

	s.Equal(http.StatusOK, s.get("/saml/metadata?domain=idp.example.com").Code)
}

func (s *samlSuite) TestMetadata_CertOnly() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	s.createCertificate()

	rec := s.get("/saml/metadata?cert_only=true")

	s.Require().Equal(http.StatusOK, rec.Code, rec.Body.String())
	s.Contains(rec.Body.String(), "-----BEGIN CERTIFICATE-----")
}

func (s *samlSuite) TestMetadata_NoCertificate() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	s.Equal(http.StatusNotFound, s.get("/saml/metadata").Code)
	s.Equal(http.StatusNotFound, s.get("/saml/metadata?cert_only=true").Code)
}
