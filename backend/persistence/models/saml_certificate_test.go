package models

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newKey(t *testing.T, bits int) *rsa.PrivateKey {
	key, err := rsa.GenerateKey(rand.Reader, bits)
	require.NoError(t, err)
	return key
}

func TestNewSamlCertificate_StoresPrivateKeyEncrypted(t *testing.T) {
	key := newKey(t, 2048)

	cert, err := NewSamlCertificate("cert data", key)
	require.NoError(t, err)

	assert.Equal(t, "cert data", cert.CertData)
	assert.NotContains(t, cert.CertKey, "PRIVATE KEY")

	stored, err := cert.PrivateKey()
	require.NoError(t, err)
	assert.True(t, key.Equal(stored))

	bits, err := cert.KeyBits()
	require.NoError(t, err)
	assert.Equal(t, 2048, bits)
}

func TestSetPrivateKey_KeepsEncryptionKey(t *testing.T) {
	cert, err := NewSamlCertificate("cert data", newKey(t, 2048))
	require.NoError(t, err)
	encryptionKey := cert.EncryptionKey
	encryptedKey := cert.CertKey

	replacement := newKey(t, 3072)
	require.NoError(t, cert.SetPrivateKey(replacement))

	assert.Equal(t, encryptionKey, cert.EncryptionKey)
	assert.NotEqual(t, encryptedKey, cert.CertKey)
	stored, err := cert.PrivateKey()
	require.NoError(t, err)
	assert.True(t, replacement.Equal(stored))
}
