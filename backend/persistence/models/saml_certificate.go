package models

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/gobuffalo/validate/v3/validators"
	"github.com/teamhanko/hanko/backend/v3/crypto"
	"github.com/teamhanko/hanko/backend/v3/crypto/aes_gcm"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
)

type SamlCertificate struct {
	ID            uuid.UUID `json:"id" db:"id"`
	TenantID      uuid.UUID `db:"tenant_id"`
	CertData      string    `json:"cert_data" db:"cert_data"`
	CertKey       string    `json:"cert_key" db:"cert_key"`
	EncryptionKey string    `json:"encryption_key" db:"encryption_key"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

func encryptPrivateKey(privateKey []byte, encryptionKey string) (string, error) {
	gcm, err := aes_gcm.NewAESGCM([]string{encryptionKey})
	if err != nil {
		return "", err
	}

	return gcm.Encrypt(privateKey)
}

// NewSamlCertificate creates a certificate from its generated certificate and private key. The private key is
// stored encrypted with a newly generated encryption key.
func NewSamlCertificate(certData string, privateKey *rsa.PrivateKey) (*SamlCertificate, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("could not generate id: %w", err)
	}

	encryptionKey, err := crypto.GenerateRandomStringURLSafe(32)
	if err != nil {
		return nil, fmt.Errorf("unable to create encryptionKey: %w", err)
	}

	encryptedPrivateKey, err := encryptPrivateKey(encodePrivateKey(privateKey), encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("unable to encrypt private key: %w", err)
	}

	now := time.Now()

	return &SamlCertificate{
		ID:            id,
		CertData:      certData,
		CertKey:       encryptedPrivateKey,
		EncryptionKey: encryptionKey,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, nil
}

func encodePrivateKey(privateKey *rsa.PrivateKey) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
}

// PrivateKey returns the decrypted private key.
func (s *SamlCertificate) PrivateKey() (*rsa.PrivateKey, error) {
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

// SetPrivateKey replaces the private key, encrypting it with the existing encryption key.
func (s *SamlCertificate) SetPrivateKey(privateKey *rsa.PrivateKey) error {
	encryptedKey, err := encryptPrivateKey(encodePrivateKey(privateKey), s.EncryptionKey)
	if err != nil {
		return fmt.Errorf("unable to encrypt private key: %w", err)
	}

	s.CertKey = encryptedKey

	return nil
}

// KeyBits returns the size of the certificate's RSA key.
func (s *SamlCertificate) KeyBits() (int, error) {
	privateKey, err := s.PrivateKey()
	if err != nil {
		return 0, err
	}

	return privateKey.N.BitLen(), nil
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
