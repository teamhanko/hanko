package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v3/dto"
	"github.com/teamhanko/hanko/backend/v3/saml"
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
	cert, err := saml.NewCertificateService(s.Storage).Generate("test", saml.CertificateParams{})
	s.Require().NoError(err)
	cert.TenantID = uuid.Must(uuid.FromString(samlCertificateTenantID))

	privateKey, err := cert.PrivateKey()
	s.Require().NoError(err)

	cert.CertData, err = saml.SelfSign("test", privateKey, saml.DefaultValidityYears, time.Now().AddDate(-11, 0, 0))
	s.Require().NoError(err)
	s.Require().NoError(s.Storage.GetSamlCertificatePersister().Create(cert))
}

func (s *samlCertificateSuite) renew(body string) *httptest.ResponseRecorder {
	cfg := test.DefaultConfig
	e := NewManagementRouter(&cfg, s.Storage)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/tenants/%s/saml/certificate/renew", samlCertificateTenantID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func (s *samlCertificateSuite) TestRenew() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	s.createExpiredCertificate()

	rec := s.renew("")

	s.Require().Equal(http.StatusOK, rec.Code, rec.Body.String())
	var renewed dto.SamlCertificateResponse
	s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &renewed), rec.Body.String())
	s.True(renewed.NotAfter.After(time.Now()))
	s.Equal(2048, renewed.KeyBits)
	s.False(renewed.KeyChanged, "the key is kept unless a different size or rotation is requested")

	// the new certificate must have been persisted, not just returned
	stored, err := s.Storage.GetSamlCertificatePersister().GetFirst(uuid.Must(uuid.FromString(samlCertificateTenantID)))
	s.Require().NoError(err)
	storedValidity, err := certificateResponse(stored)
	s.Require().NoError(err)
	s.True(storedValidity.NotAfter.Equal(renewed.NotAfter))
}

func (s *samlCertificateSuite) TestRenew_KeySizeAndValidity() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	s.createExpiredCertificate()

	rec := s.renew(`{"key_bits": 3072, "validity_years": 3}`)

	s.Require().Equal(http.StatusOK, rec.Code, rec.Body.String())
	var renewed dto.SamlCertificateResponse
	s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &renewed), rec.Body.String())
	s.True(renewed.KeyChanged)
	s.Equal(3072, renewed.KeyBits)
	s.True(renewed.NotBefore.AddDate(3, 0, 0).Equal(renewed.NotAfter))

	// the new key must have been persisted, not just returned
	stored, err := s.Storage.GetSamlCertificatePersister().GetFirst(uuid.Must(uuid.FromString(samlCertificateTenantID)))
	s.Require().NoError(err)
	storedBits, err := stored.KeyBits()
	s.Require().NoError(err)
	s.Equal(3072, storedBits)
}

func (s *samlCertificateSuite) TestRenew_InvalidRequest() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	s.createExpiredCertificate()

	tests := []struct {
		name string
		body string
	}{
		{"key size not allowed", `{"key_bits": 1024}`},
		{"validity below minimum", `{"validity_years": 0}`},
		{"validity above maximum", `{"validity_years": 11}`},
		{"malformed body", `{"key_bits": `},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(http.StatusBadRequest, s.renew(tt.body).Code)
		})
	}
}

func (s *samlCertificateSuite) TestRenew_NoCertificate() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}

	s.Equal(http.StatusNotFound, s.renew("").Code)
}
