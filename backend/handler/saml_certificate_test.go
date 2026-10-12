package handler

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v3/dto"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
	"github.com/teamhanko/hanko/backend/v3/test"
)

func TestSamlCertificateSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(samlCertificateSuite))
}

type samlCertificateSuite struct {
	test.Suite
}

// samlCertificateTenantID is the tenant inserted by the add_tenant_id migration.
const samlCertificateTenantID = "00000000-0000-0000-0000-000000000001"

// createExpiredCertificate stores a certificate that expired a year ago.
func (s *samlCertificateSuite) createExpiredCertificate() {
	cert, err := models.NewSamlCertificate("test")
	s.Require().NoError(err)
	cert.TenantID = uuid.Must(uuid.FromString(samlCertificateTenantID))

	keyPEM, err := cert.DecryptCertKey()
	s.Require().NoError(err)
	keyBlock, _ := pem.Decode(keyPEM)
	s.Require().NotNil(keyBlock)
	privateKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	s.Require().NoError(err)

	cert.CertData, err = models.GenerateCertificate("test", privateKey, time.Now().AddDate(-11, 0, 0))
	s.Require().NoError(err)
	s.Require().NoError(s.Storage.GetSamlCertificatePersister().Create(cert))
}

func (s *samlCertificateSuite) request(method string) *httptest.ResponseRecorder {
	cfg := test.DefaultConfig
	e := NewManagementRouter(&cfg, s.Storage)

	path := fmt.Sprintf("/tenants/%s/saml/certificate", samlCertificateTenantID)
	if method == http.MethodPost {
		path += "/renew"
	}

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
	return rec
}

func (s *samlCertificateSuite) decode(rec *httptest.ResponseRecorder) dto.SamlCertificateResponse {
	var response dto.SamlCertificateResponse
	s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &response), rec.Body.String())
	return response
}

func (s *samlCertificateSuite) TestGet() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	s.createExpiredCertificate()

	rec := s.request(http.MethodGet)

	s.Require().Equal(http.StatusOK, rec.Code, rec.Body.String())
	s.True(s.decode(rec).NotAfter.Before(time.Now()))
}

func (s *samlCertificateSuite) TestRenew() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	s.createExpiredCertificate()

	rec := s.request(http.MethodPost)

	s.Require().Equal(http.StatusOK, rec.Code, rec.Body.String())
	renewed := s.decode(rec)
	s.True(renewed.NotAfter.After(time.Now()))

	// the new certificate must have been persisted, not just returned
	s.True(s.decode(s.request(http.MethodGet)).NotAfter.Equal(renewed.NotAfter))
}

func (s *samlCertificateSuite) TestNoCertificate() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	s.Equal(http.StatusNotFound, s.request(http.MethodGet).Code)
	s.Equal(http.StatusNotFound, s.request(http.MethodPost).Code)
}
