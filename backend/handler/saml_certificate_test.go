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

func (s *samlCertificateSuite) renew() *httptest.ResponseRecorder {
	cfg := test.DefaultConfig
	e := NewManagementRouter(&cfg, s.Storage)

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, fmt.Sprintf("/tenants/%s/saml/certificate/renew", samlCertificateTenantID), nil))
	return rec
}

func (s *samlCertificateSuite) TestRenew() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	s.createExpiredCertificate()

	rec := s.renew()

	s.Require().Equal(http.StatusOK, rec.Code, rec.Body.String())
	var renewed dto.SamlCertificateResponse
	s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &renewed), rec.Body.String())
	s.True(renewed.NotAfter.After(time.Now()))

	// the new certificate must have been persisted, not just returned
	stored, err := s.Storage.GetSamlCertificatePersister().GetFirst(uuid.Must(uuid.FromString(samlCertificateTenantID)))
	s.Require().NoError(err)
	storedValidity, err := certificateResponse(stored)
	s.Require().NoError(err)
	s.True(storedValidity.NotAfter.Equal(renewed.NotAfter))
}

func (s *samlCertificateSuite) TestRenew_NoCertificate() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	s.Equal(http.StatusNotFound, s.renew().Code)
}
