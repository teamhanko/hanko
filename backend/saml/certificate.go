package saml

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"
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

// maxCommonNameLength is the upper bound for a common name (RFC 5280).
const maxCommonNameLength = 64

// CertificateSubject is the identity of a SAML certificate: its common name and subject alternative names.
type CertificateSubject struct {
	CommonName string
	DNSNames   []string
}

// NewCertificateSubject derives the subject from the host of the SAML endpoint URL, which becomes the common name
// and the only subject alternative name.
//
// A common name is limited to 64 characters, so a longer host is only set as the subject alternative name and the
// service name stays the common name. Without a usable host, which is none at all, an IP address or one that can't
// be encoded in a subject alternative name (a non-ASCII one), the service name is the common name and there is no
// subject alternative name. The service name is never used as a subject alternative name because it is free text
// and not a host name.
func NewCertificateSubject(serviceName, endpoint string) CertificateSubject {
	subject := CertificateSubject{CommonName: serviceName}

	host := endpointHost(endpoint)
	if host == "" {
		return subject
	}

	subject.DNSNames = []string{host}
	if len(host) <= maxCommonNameLength {
		subject.CommonName = host
	}

	return subject
}

func (s CertificateSubject) String() string {
	if len(s.DNSNames) == 0 {
		return fmt.Sprintf("CN=%q, no SAN", s.CommonName)
	}

	return fmt.Sprintf("CN=%q, SAN=%s", s.CommonName, strings.Join(s.DNSNames, ","))
}

// endpointHost returns the lower-cased host name of the endpoint URL without port, or an empty string if there is
// no host name that can be used.
func endpointHost(endpoint string) string {
	u, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil {
		return ""
	}

	host := strings.ToLower(u.Hostname())
	if net.ParseIP(host) != nil {
		return ""
	}

	for i := 0; i < len(host); i++ {
		if host[i] >= utf8.RuneSelf {
			return ""
		}
	}

	return host
}

// SelfSign creates a PEM encoded certificate for the key, signed by the key itself, which is valid from now for
// the given number of whole years.
func SelfSign(subject CertificateSubject, key *rsa.PrivateKey, validityYears int, now time.Time) (string, error) {
	if err := ValidateValidityYears(validityYears); err != nil {
		return "", err
	}

	// SerialNumber is left unset, so CreateCertificate generates a random one. The certificate is self-signed, so
	// issuer and subject are the same and a fixed serial would repeat on every renewal.
	template := &x509.Certificate{
		Subject: pkix.Name{
			CommonName: subject.CommonName,
		},
		DNSNames:              subject.DNSNames,
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
