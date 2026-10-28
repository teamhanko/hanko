package certificate

import (
	"testing"

	"github.com/teamhanko/hanko/backend/v3/saml"
)

func (s *certificateSuite) TestCreate_CreatesCertificateForTenant() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	tenantID := s.createTenant(tenantConfigWithEndpoint)

	result, err := createCertificateForTenant(withMultiTenancy, s.Storage, tenantID, false, certificateParams{keyBits: 3072, validityYears: 3})

	s.Require().NoError(err)
	s.Equal(outcomeCreated, result.Outcome)
	cert := s.storedCertificate(tenantID)
	s.Equal(3072, s.keyBits(cert))
	parsed := s.parsed(cert)
	s.True(parsed.NotBefore.AddDate(3, 0, 0).Equal(parsed.NotAfter))
	s.Equal("auth.example.com", parsed.Subject.CommonName)
	s.Equal([]string{"auth.example.com"}, parsed.DNSNames)
}

func (s *certificateSuite) TestCreate_DryRunWritesNothing() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	tenantID := s.createTenant(tenantConfigWithEndpoint)

	result, err := createCertificateForTenant(withMultiTenancy, s.Storage, tenantID, true, flagDefaults)

	s.Require().NoError(err)
	s.Equal(outcomeWouldCreate, result.Outcome)
	s.Contains(result.Message, `CN="auth.example.com", SAN=auth.example.com`)
	stored, err := s.Storage.GetSamlCertificatePersister().GetFirst(tenantID)
	s.Require().NoError(err)
	s.Nil(stored)
}

// create-all can be run again and again, which relies on a tenant's certificate being left alone.
func (s *certificateSuite) TestCreate_KeepsExistingCertificate() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	tenantID := s.createTenant(tenantConfigWithEndpoint)
	existing := s.storeCertificate(tenantID, saml.CertificateParams{})

	result, err := createCertificateForTenant(withMultiTenancy, s.Storage, tenantID, false, flagDefaults)

	s.Require().NoError(err)
	s.Equal(outcomeAlreadyExists, result.Outcome)
	s.Equal(existing.CertData, s.storedCertificate(tenantID).CertData)
}

func (s *certificateSuite) TestCreate_FailsForTenantWithoutUsableName() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	tenantID := s.createTenant(`{"service": {"name": ""}}`)

	_, err := createCertificateForTenant(withMultiTenancy, s.Storage, tenantID, false, flagDefaults)

	s.Require().Error(err)
	stored, err := s.Storage.GetSamlCertificatePersister().GetFirst(tenantID)
	s.Require().NoError(err)
	s.Nil(stored)
}
