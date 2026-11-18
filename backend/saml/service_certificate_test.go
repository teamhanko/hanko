package saml

import (
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

// The service methods under test don't use the persister, only Renew persists.
var certificateService = NewCertificateService(nil)

func generateCertificate(t *testing.T, params CertificateParams) *models.SamlCertificate {
	cert, err := certificateService.Generate(testSubject, params)
	require.NoError(t, err)
	return cert
}

func TestGenerate_Defaults(t *testing.T) {
	cert := generateCertificate(t, CertificateParams{})

	bits, err := cert.KeyBits()
	require.NoError(t, err)
	assert.Equal(t, DefaultKeyBits, bits)

	parsed := parseCert(t, cert.CertData)
	assert.True(t, parsed.NotBefore.AddDate(DefaultValidityYears, 0, 0).Equal(parsed.NotAfter))
}

func TestGenerate_Params(t *testing.T) {
	cert := generateCertificate(t, CertificateParams{KeyBits: 3072, ValidityYears: 3})

	bits, err := cert.KeyBits()
	require.NoError(t, err)
	assert.Equal(t, 3072, bits)

	parsed := parseCert(t, cert.CertData)
	assert.True(t, parsed.NotBefore.AddDate(3, 0, 0).Equal(parsed.NotAfter))
	assert.Equal(t, 3072, parsed.PublicKey.(*rsa.PublicKey).N.BitLen())
}

func TestGenerate_InvalidParams(t *testing.T) {
	tests := []struct {
		name   string
		params CertificateParams
	}{
		{"key size too small", CertificateParams{KeyBits: 1024}},
		{"key size not allowed", CertificateParams{KeyBits: 2000}},
		{"validity below minimum", CertificateParams{ValidityYears: -1}},
		{"validity above maximum", CertificateParams{ValidityYears: 11}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := certificateService.Generate(testSubject, tt.params)
			assert.Error(t, err)
		})
	}
}

func TestRenewCertificate_KeepsKeyByDefault(t *testing.T) {
	cert := generateCertificate(t, CertificateParams{})
	encryptedKey := cert.CertKey

	keyChanged, err := renewCertificate(cert, testSubject, CertificateParams{ValidityYears: 2})
	require.NoError(t, err)

	assert.False(t, keyChanged)
	assert.Equal(t, encryptedKey, cert.CertKey)
	parsed := parseCert(t, cert.CertData)
	assert.True(t, parsed.NotBefore.AddDate(2, 0, 0).Equal(parsed.NotAfter))
}

func TestRenewCertificate_SameKeySizeKeepsKey(t *testing.T) {
	cert := generateCertificate(t, CertificateParams{})
	encryptedKey := cert.CertKey

	keyChanged, err := renewCertificate(cert, testSubject, CertificateParams{KeyBits: 2048})

	require.NoError(t, err)
	assert.False(t, keyChanged)
	assert.Equal(t, encryptedKey, cert.CertKey)
}

func TestRenewCertificate_DifferentKeySizeReplacesKey(t *testing.T) {
	cert := generateCertificate(t, CertificateParams{})
	encryptedKey := cert.CertKey

	keyChanged, err := renewCertificate(cert, testSubject, CertificateParams{KeyBits: 3072})
	require.NoError(t, err)

	assert.True(t, keyChanged)
	assert.NotEqual(t, encryptedKey, cert.CertKey)
	bits, err := cert.KeyBits()
	require.NoError(t, err)
	assert.Equal(t, 3072, bits)
	assert.Equal(t, 3072, parseCert(t, cert.CertData).PublicKey.(*rsa.PublicKey).N.BitLen())
}

func TestRenewCertificate_RotateKeyKeepsSize(t *testing.T) {
	cert := generateCertificate(t, CertificateParams{})
	encryptedKey := cert.CertKey

	keyChanged, err := renewCertificate(cert, testSubject, CertificateParams{RotateKey: true})
	require.NoError(t, err)

	assert.True(t, keyChanged)
	assert.NotEqual(t, encryptedKey, cert.CertKey)
	bits, err := cert.KeyBits()
	require.NoError(t, err)
	assert.Equal(t, 2048, bits)
}

func TestRenewCertificate_GeneratesNewSerial(t *testing.T) {
	cert := generateCertificate(t, CertificateParams{})
	serial := parseCert(t, cert.CertData).SerialNumber

	_, err := renewCertificate(cert, testSubject, CertificateParams{})
	require.NoError(t, err)

	assert.NotEqual(t, serial, parseCert(t, cert.CertData).SerialNumber)
}

func TestRenewCertificate_InvalidParamsLeaveCertificateUntouched(t *testing.T) {
	cert := generateCertificate(t, CertificateParams{})
	certData, certKey := cert.CertData, cert.CertKey

	_, err := renewCertificate(cert, testSubject, CertificateParams{KeyBits: 1024})

	assert.Error(t, err)
	assert.Equal(t, certData, cert.CertData)
	assert.Equal(t, certKey, cert.CertKey)
}

func TestWouldReplaceKey(t *testing.T) {
	cert := generateCertificate(t, CertificateParams{})

	tests := []struct {
		name   string
		params CertificateParams
		want   bool
	}{
		{"no params", CertificateParams{}, false},
		{"same size", CertificateParams{KeyBits: 2048}, false},
		{"different size", CertificateParams{KeyBits: 4096}, true},
		{"rotate", CertificateParams{RotateKey: true}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := certificateService.WouldReplaceKey(cert, tt.params)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
