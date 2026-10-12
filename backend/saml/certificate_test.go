package saml

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
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

	certPEM, err := SelfSign("test", key, 3, now)
	require.NoError(t, err)

	cert := parseCert(t, certPEM)
	assert.Equal(t, "test", cert.Subject.CommonName)
	assert.True(t, now.AddDate(3, 0, 0).Truncate(time.Second).Equal(cert.NotAfter))
	assert.True(t, key.PublicKey.Equal(cert.PublicKey.(*rsa.PublicKey)))
	assert.NoError(t, cert.CheckSignature(cert.SignatureAlgorithm, cert.RawTBSCertificate, cert.Signature), "must be signed by its own key")
}

func TestSelfSign_GeneratesNewSerialEachTime(t *testing.T) {
	key, err := GenerateKey(DefaultKeyBits)
	require.NoError(t, err)

	first, err := SelfSign("test", key, 1, time.Now())
	require.NoError(t, err)
	second, err := SelfSign("test", key, 1, time.Now())
	require.NoError(t, err)

	serial := parseCert(t, first).SerialNumber
	assert.NotEqual(t, serial, parseCert(t, second).SerialNumber)
	assert.Equal(t, 1, serial.Sign(), "serial must be positive")
}

func TestSelfSign_InvalidValidity(t *testing.T) {
	key, err := GenerateKey(DefaultKeyBits)
	require.NoError(t, err)

	_, err = SelfSign("test", key, 0, time.Now())
	assert.Error(t, err)
}
