package saml

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"time"
)

const (
	DefaultKeyBits       = 2048
	DefaultValidityYears = 10
	MinValidityYears     = 1
	MaxValidityYears     = 10
)

// AllowedKeyBits are the RSA key sizes certificates can be created or renewed with.
var AllowedKeyBits = []int{2048, 3072, 4096}

// ValidateKeyBits reports whether bits is an allowed RSA key size.
func ValidateKeyBits(bits int) error {
	for _, allowed := range AllowedKeyBits {
		if bits == allowed {
			return nil
		}
	}
	return fmt.Errorf("key size must be one of %v bits, got %d", AllowedKeyBits, bits)
}

// ValidateValidityYears reports whether years is an allowed certificate validity.
func ValidateValidityYears(years int) error {
	if years < MinValidityYears || years > MaxValidityYears {
		return fmt.Errorf("validity must be between %d and %d years, got %d", MinValidityYears, MaxValidityYears, years)
	}
	return nil
}

// GenerateKey generates an RSA key of one of the AllowedKeyBits sizes.
func GenerateKey(bits int) (*rsa.PrivateKey, error) {
	if err := ValidateKeyBits(bits); err != nil {
		return nil, err
	}

	key, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, fmt.Errorf("unable to generate private key: %w", err)
	}

	return key, nil
}

// SelfSign creates a PEM encoded certificate for the key, signed by the key itself, which is valid from now for
// the given number of whole years.
func SelfSign(serviceName string, key *rsa.PrivateKey, validityYears int, now time.Time) (string, error) {
	if err := ValidateValidityYears(validityYears); err != nil {
		return "", err
	}

	// SerialNumber is left unset, so CreateCertificate generates a random one. The certificate is self-signed, so
	// issuer and subject are the same and a fixed serial would repeat on every renewal.
	template := &x509.Certificate{
		Subject: pkix.Name{
			CommonName: serviceName,
		},
		NotBefore:             now,
		NotAfter:              now.AddDate(validityYears, 0, 0),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	cert, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return "", fmt.Errorf("unable to create certificate: %w", err)
	}

	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert})), nil
}
