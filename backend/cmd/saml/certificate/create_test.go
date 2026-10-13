package certificate

import (
	"testing"
	"time"

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

// create-all calls createCertificateForTenant concurrently, and a tenant must still end up with one certificate.
func (s *certificateSuite) TestCreate_ConcurrentCallsCreateOneCertificate() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	tenantID := s.createTenant(tenantConfigWithEndpoint)

	// The calls only race for the certificate in the short moment between checking and inserting. Holding the
	// tenant's row lock until all of them are waiting for it makes them arrive there at the same time, which is
	// what they would do if the lock didn't exist.
	tx, err := s.Storage.GetConnection().NewTransaction()
	s.Require().NoError(err)
	s.Require().NoError(tx.RawQuery("SELECT id FROM tenants WHERE id = ? FOR UPDATE", tenantID).Exec())

	const calls = 5
	results := make(chan certResult, calls)
	errs := make(chan error, calls)
	for range calls {
		go func() {
			result, err := createCertificateForTenant(withMultiTenancy, s.Storage, tenantID, false, flagDefaults)
			results <- result
			errs <- err
		}()
	}

	s.Eventually(func() bool {
		var waiting struct {
			Count int `db:"count"`
		}
		err := s.Storage.GetConnection().RawQuery("SELECT count(*) AS count FROM pg_stat_activity WHERE wait_event_type = 'Lock'").First(&waiting)
		s.Require().NoError(err)
		return waiting.Count == calls
	}, 10*time.Second, 20*time.Millisecond, "all calls must be waiting for the tenant's row lock")
	s.Require().NoError(tx.TX.Commit())

	created := 0
	for range calls {
		s.Require().NoError(<-errs)
		if (<-results).Outcome == outcomeCreated {
			created++
		}
	}

	s.Equal(1, created)
}
