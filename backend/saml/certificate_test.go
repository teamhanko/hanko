package saml

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func parseCert(t *testing.T, certPEM string) *x509.Certificate {
	block, _ := pem.Decode([]byte(certPEM))
	require.NotNil(t, block)
	cert, err := x509.ParseCertificate(block.Bytes)
	require.NoError(t, err)
	return cert
}

// testSubject has no endpoint, so its common name is the service name and there is no SAN.
var testSubject = NewCertificateSubject("test", "")

func TestNewCertificateSubject(t *testing.T) {
	longHost := strings.Repeat("a", 60) + ".example.com"

	tests := []struct {
		name     string
		endpoint string
		wantCN   string
		wantDNS  []string
	}{
		{"https endpoint", "https://auth.example.com", "auth.example.com", []string{"auth.example.com"}},
		{"trailing slashes", "https://auth.example.com///", "auth.example.com", []string{"auth.example.com"}},
		{"path", "https://hanko.example.com/api", "hanko.example.com", []string{"hanko.example.com"}},
		{"port", "http://localhost:8000", "localhost", []string{"localhost"}},
		{"upper case", "https://Auth.Example.COM", "auth.example.com", []string{"auth.example.com"}},
		{"empty endpoint", "", "My App", nil},
		{"no scheme", "auth.example.com", "My App", nil},
		{"ip address", "http://10.0.0.5:8000", "My App", nil},
		{"non-ASCII host", "https://münchen.example.com", "My App", nil},
		{"unparseable", "https://exa mple.com", "My App", nil},
		{"host too long for a common name", "https://" + longHost, "My App", []string{longHost}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subject := NewCertificateSubject("My App", tt.endpoint)

			assert.Equal(t, tt.wantCN, subject.CommonName)
			assert.Equal(t, tt.wantDNS, subject.DNSNames)
		})
	}
}

func TestValidateKeyBits(t *testing.T) {
	for _, bits := range AllowedKeyBits {
		assert.NoError(t, ValidateKeyBits(bits))
	}
	for _, bits := range []int{0, 1024, 2000, 8192} {
		assert.Error(t, ValidateKeyBits(bits), "%d bits", bits)
	}
}

func TestValidateValidityYears(t *testing.T) {
	for years := MinValidityYears; years <= MaxValidityYears; years++ {
		assert.NoError(t, ValidateValidityYears(years))
	}
	for _, years := range []int{-1, 0, MaxValidityYears + 1} {
		assert.Error(t, ValidateValidityYears(years), "%d years", years)
	}
}

func TestGenerateKey(t *testing.T) {
	key, err := GenerateKey(3072)
	require.NoError(t, err)
	assert.Equal(t, 3072, key.N.BitLen())

	_, err = GenerateKey(1024)
	assert.Error(t, err)
}

func TestSelfSign(t *testing.T) {
	key, err := GenerateKey(DefaultKeyBits)
	require.NoError(t, err)
	now := time.Now()

	certPEM, err := SelfSign(NewCertificateSubject("test", "https://auth.example.com/"), key, 3, now)
	require.NoError(t, err)

	cert := parseCert(t, certPEM)
	assert.Equal(t, "auth.example.com", cert.Subject.CommonName)
	assert.Equal(t, []string{"auth.example.com"}, cert.DNSNames)
	assert.True(t, now.AddDate(3, 0, 0).Truncate(time.Second).Equal(cert.NotAfter))
	assert.True(t, key.PublicKey.Equal(cert.PublicKey.(*rsa.PublicKey)))
	assert.NoError(t, cert.CheckSignature(cert.SignatureAlgorithm, cert.RawTBSCertificate, cert.Signature), "must be signed by its own key")
}

func TestSelfSign_GeneratesNewSerialEachTime(t *testing.T) {
	key, err := GenerateKey(DefaultKeyBits)
	require.NoError(t, err)

	first, err := SelfSign(testSubject, key, 1, time.Now())
	require.NoError(t, err)
	second, err := SelfSign(testSubject, key, 1, time.Now())
	require.NoError(t, err)

	serial := parseCert(t, first).SerialNumber
	assert.NotEqual(t, serial, parseCert(t, second).SerialNumber)
	assert.Equal(t, 1, serial.Sign(), "serial must be positive")
}

func TestSelfSign_InvalidValidity(t *testing.T) {
	key, err := GenerateKey(DefaultKeyBits)
	require.NoError(t, err)

	_, err = SelfSign(testSubject, key, 0, time.Now())
	assert.Error(t, err)
}
