package certificate

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamhanko/hanko/backend/v3/saml"
)

func TestRenewCommand_Args(t *testing.T) {
	tenantID := uuid.Must(uuid.NewV4()).String()

	tests := []struct {
		name    string
		flags   map[string]string
		wantErr bool
	}{
		{"tenant only", map[string]string{"tenant_id": tenantID}, false},
		{"key size and validity", map[string]string{"tenant_id": tenantID, "key-bits": "4096", "validity-years": "3"}, false},
		{"no tenant", map[string]string{}, true},
		{"tenant that isn't a uuid", map[string]string{"tenant_id": "tenant"}, true},
		{"key size that isn't allowed", map[string]string{"tenant_id": tenantID, "key-bits": "1024"}, true},
		// without --key-bits the current key size is kept, but 0 given explicitly is not that
		{"key size 0 given explicitly", map[string]string{"tenant_id": tenantID, "key-bits": "0"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewRenewCommand()
			for name, value := range tt.flags {
				require.NoError(t, cmd.Flags().Set(name, value))
			}

			err := cmd.Args(cmd, nil)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// renewParams are the parameters of a renewal that only asks for a validity, like the command without any other flag.
var renewParams = certificateParams{validityYears: 5}

func (s *certificateSuite) TestRenew_ReissuesCertificateAndKeepsKey() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	tenantID := s.createTenant(tenantConfigWithEndpoint)
	before := s.storeCertificate(tenantID, saml.CertificateParams{})

	result, err := renewCertificateForTenant(withMultiTenancy, s.Storage, tenantID, false, renewParams)

	s.Require().NoError(err)
	s.Equal(outcomeRenewed, result.Outcome)
	s.Contains(result.Message, "key pair kept")
	after := s.storedCertificate(tenantID)
	s.Equal(before.CertKey, after.CertKey, "the key must be kept")
	s.NotEqual(s.parsed(before).SerialNumber, s.parsed(after).SerialNumber)
	parsed := s.parsed(after)
	s.True(parsed.NotBefore.AddDate(5, 0, 0).Equal(parsed.NotAfter))
	s.Equal("auth.example.com", parsed.Subject.CommonName, "the subject must follow the tenant's endpoint")
	s.Equal([]string{"auth.example.com"}, parsed.DNSNames)
}

func (s *certificateSuite) TestRenew_WithoutKeySizeKeepsTheCurrentKeySize() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	tenantID := s.createTenant(tenantConfigWithEndpoint)
	before := s.storeCertificate(tenantID, saml.CertificateParams{KeyBits: 3072})

	_, err := renewCertificateForTenant(withMultiTenancy, s.Storage, tenantID, false, renewParams)

	s.Require().NoError(err)
	after := s.storedCertificate(tenantID)
	s.Equal(3072, s.keyBits(after))
	s.Equal(before.CertKey, after.CertKey)
}

func (s *certificateSuite) TestRenew_DifferentKeySizeReplacesKey() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	tenantID := s.createTenant(tenantConfigWithEndpoint)
	before := s.storeCertificate(tenantID, saml.CertificateParams{})

	result, err := renewCertificateForTenant(withMultiTenancy, s.Storage, tenantID, false, certificateParams{keyBits: 3072, validityYears: 5})

	s.Require().NoError(err)
	s.Equal(outcomeRenewed, result.Outcome)
	s.Contains(result.Message, "replaced with a new 3072-bit key")
	s.Contains(result.Message, "the key pair changed")
	after := s.storedCertificate(tenantID)
	s.NotEqual(before.CertKey, after.CertKey)
	s.Equal(3072, s.keyBits(after))
}

func (s *certificateSuite) TestRenew_DryRunChangesNothing() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	tenantID := s.createTenant(tenantConfigWithEndpoint)
	before := s.storeCertificate(tenantID, saml.CertificateParams{})

	result, err := renewCertificateForTenant(withMultiTenancy, s.Storage, tenantID, true, certificateParams{keyBits: 4096, validityYears: 5})

	s.Require().NoError(err)
	s.Equal(outcomeWouldRenew, result.Outcome)
	s.Contains(result.Message, `CN="auth.example.com", SAN=auth.example.com`)
	s.Contains(result.Message, "replaced with a new 4096-bit key")
	after := s.storedCertificate(tenantID)
	s.Equal(before.CertData, after.CertData)
	s.Equal(before.CertKey, after.CertKey)
}

func (s *certificateSuite) TestRenew_FailsForTenantWithoutUsableNameAndKeepsCertificate() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode.")
	}
	tenantID := s.createTenant(`{"service": {"name": ""}}`)
	before := s.storeCertificate(tenantID, saml.CertificateParams{})

	_, err := renewCertificateForTenant(withMultiTenancy, s.Storage, tenantID, false, renewParams)

	s.Require().Error(err)
	s.Equal(before.CertData, s.storedCertificate(tenantID).CertData)
}
