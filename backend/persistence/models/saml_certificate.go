package models

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/teamhanko/hanko/backend/v3/crypto"
	"github.com/teamhanko/hanko/backend/v3/crypto/aes_gcm"
	"math/big"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
)

// SamlCertificate is used by pop to map your saml_certs database table to your go code.
type SamlCertificate struct {
	ID            uuid.UUID `json:"id" db:"id"`
	TenantID      uuid.UUID `db:"tenant_id"`
	CertData      string    `json:"cert_data" db:"cert_data"`
	CertKey       string    `json:"cert_key" db:"cert_key"`
	EncryptionKey string    `json:"encryption_key" db:"encryption_key"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

func createTemplate(serviceName string, creationTime time.Time, validityYears int) *x509.Certificate {
	return &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: serviceName,
		},
		NotBefore:             creationTime,
		NotAfter:              creationTime.AddDate(validityYears, 0, 0),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}
}

func GenerateCertificate(serviceName string, privateKey *rsa.PrivateKey, currentTime time.Time, opts ...CertificateOption) (string, error) {
	options, err := newCertificateOptions(opts)
	if err != nil {
		return "", err
	}

	template := createTemplate(serviceName, currentTime, options.validityYears)
	cert, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return "", err
	}

	certPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	})

	return string(certPem), nil
}

func encryptPrivateKey(privateKey []byte, encryptionKey string) (string, error) {
	gcm, err := aes_gcm.NewAESGCM([]string{encryptionKey})
	if err != nil {
		return "", err
	}
	encryptedKey, err := gcm.Encrypt(privateKey)

	return encryptedKey, nil
}

func NewSamlCertificate(serviceName string, opts ...CertificateOption) (*SamlCertificate, error) {
	options, err := newCertificateOptions(opts)
	if err != nil {
		return nil, err
	}

	id, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("could not generate id: %w", err)
	}

	now := time.Now()

	privateKey, err := rsa.GenerateKey(rand.Reader, options.keyBitsOrDefault())
	if err != nil {
		return nil, fmt.Errorf("unable to generate private key: %w", err)
	}

	privateKeyPEM := x509.MarshalPKCS1PrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("error encoding private key: %w", err)

	}
	privateKeyPEMBlock := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privateKeyPEM})

	encryptionKey, err := crypto.GenerateRandomStringURLSafe(32)
	if err != nil {
		return nil, fmt.Errorf("unable to create encryptionKey: %w", err)
	}

	encryptedPrivateKey, err := encryptPrivateKey(privateKeyPEMBlock, encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("unable to encrypt private key: %w", err)
	}

	cert, err := GenerateCertificate(serviceName, privateKey, now, opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to create certificate: %w", err)
	}

	return &SamlCertificate{
		ID:            id,
		CertData:      cert,
		CertKey:       encryptedPrivateKey,
		EncryptionKey: encryptionKey,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

func (s *SamlCertificate) privateKey() (*rsa.PrivateKey, error) {
	keyPEM, err := s.DecryptCertKey()
	if err != nil {
		return nil, fmt.Errorf("unable to decrypt private key: %w", err)
	}

	decodedKey, _ := pem.Decode(keyPEM)
	if decodedKey == nil || decodedKey.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("unable to decode private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(decodedKey.Bytes)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %w", err)
	}

	return privateKey, nil
}

// KeyBits returns the size of the certificate's RSA key.
func (s *SamlCertificate) KeyBits() (int, error) {
	privateKey, err := s.privateKey()
	if err != nil {
		return 0, err
	}

	return privateKey.N.BitLen(), nil
}

// WouldReplaceKey reports whether renewing with opts generates a new key pair.
func (s *SamlCertificate) WouldReplaceKey(opts ...CertificateOption) (bool, error) {
	options, err := newCertificateOptions(opts)
	if err != nil {
		return false, err
	}

	currentBits, err := s.KeyBits()
	if err != nil {
		return false, err
	}

	return options.replacesKey(currentBits), nil
}

// Renew reissues the certificate with a new validity period. The key pair is kept unless a key size other than
// the current one is requested or the key is explicitly rotated, in which case a new key pair is generated and the
// private key is re-encrypted. It reports whether the key pair was replaced.
func (s *SamlCertificate) Renew(serviceName string, opts ...CertificateOption) (bool, error) {
	options, err := newCertificateOptions(opts)
	if err != nil {
		return false, err
	}

	privateKey, err := s.privateKey()
	if err != nil {
		return false, err
	}

	currentBits := privateKey.N.BitLen()
	keyChanged := options.replacesKey(currentBits)

	var encryptedKey string
	if keyChanged {
		bits := options.keyBits
		if bits == 0 {
			bits = currentBits
		}

		privateKey, err = rsa.GenerateKey(rand.Reader, bits)
		if err != nil {
			return false, fmt.Errorf("unable to generate private key: %w", err)
		}

		privateKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
		encryptedKey, err = encryptPrivateKey(privateKeyPEM, s.EncryptionKey)
		if err != nil {
			return false, fmt.Errorf("unable to encrypt private key: %w", err)
		}
	}

	certData, err := GenerateCertificate(serviceName, privateKey, time.Now(), opts...)
	if err != nil {
		return false, fmt.Errorf("unable to renew certificate: %w", err)
	}

	s.CertData = certData
	if keyChanged {
		s.CertKey = encryptedKey
	}

	return keyChanged, nil
}

func (s *SamlCertificate) DecryptCertKey() ([]byte, error) {
	gcm, err := aes_gcm.NewAESGCM([]string{s.EncryptionKey})
	if err != nil {
		return nil, err
	}
	encryptedKey, err := gcm.Decrypt(s.CertKey)
	if err != nil {
		return nil, err
	}

	return encryptedKey, nil
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (s *SamlCertificate) Validate(_ *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.UUIDIsPresent{Name: "ID", Field: s.ID},
		&validators.StringIsPresent{Name: "CertData", Field: s.CertData},
		&validators.StringIsPresent{Name: "CertKey", Field: s.CertKey},
		&validators.StringIsPresent{Name: "EncryptionKey", Field: s.EncryptionKey},
		&validators.StringLengthInRange{
			Name:  "EncryptionKey",
			Field: s.EncryptionKey,
			Min:   32,
		},
	), nil
}
