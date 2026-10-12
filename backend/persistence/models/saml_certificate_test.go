package models

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

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

func TestNewSamlCertificate_Defaults(t *testing.T) {
	cert, err := NewSamlCertificate("test")
	require.NoError(t, err)

	bits, err := cert.KeyBits()
	require.NoError(t, err)
	assert.Equal(t, 2048, bits)

	parsed := parseCert(t, cert.CertData)
	assert.True(t, parsed.NotBefore.AddDate(10, 0, 0).Equal(parsed.NotAfter))
}

func TestNewSamlCertificate_Options(t *testing.T) {
	cert, err := NewSamlCertificate("test", WithKeyBits(3072), WithValidityYears(3))
	require.NoError(t, err)

	bits, err := cert.KeyBits()
	require.NoError(t, err)
	assert.Equal(t, 3072, bits)

	parsed := parseCert(t, cert.CertData)
	assert.True(t, parsed.NotBefore.AddDate(3, 0, 0).Equal(parsed.NotAfter))
	assert.Equal(t, 3072, parsed.PublicKey.(*rsa.PublicKey).N.BitLen())
}

func TestNewSamlCertificate_InvalidOptions(t *testing.T) {
	tests := []struct {
		name string
		opts []CertificateOption
	}{
		{"key size too small", []CertificateOption{WithKeyBits(1024)}},
		{"key size not allowed", []CertificateOption{WithKeyBits(2000)}},
		{"validity below minimum", []CertificateOption{WithValidityYears(0)}},
		{"validity above maximum", []CertificateOption{WithValidityYears(11)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewSamlCertificate("test", tt.opts...)
			assert.Error(t, err)
		})
	}
}

func TestRenew_GeneratesNewSerial(t *testing.T) {
	cert, err := NewSamlCertificate("test")
	require.NoError(t, err)
	serial := parseCert(t, cert.CertData).SerialNumber

	_, err = cert.Renew("test")
	require.NoError(t, err)

	renewed := parseCert(t, cert.CertData).SerialNumber
	assert.NotEqual(t, serial, renewed)
	assert.Equal(t, 1, renewed.Sign(), "serial must be positive")
}

func TestRenew_KeepsKeyByDefault(t *testing.T) {
	cert, err := NewSamlCertificate("test")
	require.NoError(t, err)
	encryptedKey := cert.CertKey

	keyChanged, err := cert.Renew("test", WithValidityYears(2))
	require.NoError(t, err)

	assert.False(t, keyChanged)
	assert.Equal(t, encryptedKey, cert.CertKey)
	parsed := parseCert(t, cert.CertData)
	assert.True(t, parsed.NotBefore.AddDate(2, 0, 0).Equal(parsed.NotAfter))
}

func TestRenew_SameKeySizeKeepsKey(t *testing.T) {
	cert, err := NewSamlCertificate("test")
	require.NoError(t, err)
	encryptedKey := cert.CertKey

	keyChanged, err := cert.Renew("test", WithKeyBits(2048))

	require.NoError(t, err)
	assert.False(t, keyChanged)
	assert.Equal(t, encryptedKey, cert.CertKey)
}

func TestRenew_DifferentKeySizeReplacesKey(t *testing.T) {
	cert, err := NewSamlCertificate("test")
	require.NoError(t, err)
	encryptedKey := cert.CertKey

	keyChanged, err := cert.Renew("test", WithKeyBits(3072))
	require.NoError(t, err)

	assert.True(t, keyChanged)
	assert.NotEqual(t, encryptedKey, cert.CertKey)
	bits, err := cert.KeyBits()
	require.NoError(t, err)
	assert.Equal(t, 3072, bits)
	assert.Equal(t, 3072, parseCert(t, cert.CertData).PublicKey.(*rsa.PublicKey).N.BitLen())
}

func TestRenew_RotateKeyKeepsSize(t *testing.T) {
	cert, err := NewSamlCertificate("test")
	require.NoError(t, err)
	encryptedKey := cert.CertKey

	keyChanged, err := cert.Renew("test", WithRotateKey())
	require.NoError(t, err)

	assert.True(t, keyChanged)
	assert.NotEqual(t, encryptedKey, cert.CertKey)
	bits, err := cert.KeyBits()
	require.NoError(t, err)
	assert.Equal(t, 2048, bits)
}

func TestWouldReplaceKey(t *testing.T) {
	cert, err := NewSamlCertificate("test")
	require.NoError(t, err)

	tests := []struct {
		name string
		opts []CertificateOption
		want bool
	}{
		{"no options", nil, false},
		{"same size", []CertificateOption{WithKeyBits(2048)}, false},
		{"different size", []CertificateOption{WithKeyBits(4096)}, true},
		{"rotate", []CertificateOption{WithRotateKey()}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cert.WouldReplaceKey(tt.opts...)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
