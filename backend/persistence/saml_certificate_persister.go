package persistence

import (
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type SamlCertificatePersister interface {
	Create(cert *models.SamlCertificate) error
	GetFirst() (*models.SamlCertificate, error)
	Renew(cert *models.SamlCertificate, serviceName string) error
	Delete(cert *models.SamlCertificate) error
}

type samlCertificatePersister struct {
	db *pop.Connection
}

func NewSamlCertificatePersister(db *pop.Connection) SamlCertificatePersister {
	return &samlCertificatePersister{db: db}
}

func (s samlCertificatePersister) GetFirst() (*models.SamlCertificate, error) {
	cert := models.SamlCertificate{}

	err := s.db.First(&cert)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get certificate: %w", err)
	}

	return &cert, nil
}

func (s samlCertificatePersister) Create(cert *models.SamlCertificate) error {
	validationError, err := s.db.ValidateAndCreate(cert)
	if err != nil {
		return err
	}

	if validationError != nil && validationError.HasAny() {
		return fmt.Errorf("token object validation failed: %w", validationError)
	}

	return nil
}

func (s samlCertificatePersister) Renew(cert *models.SamlCertificate, serviceName string) error {
	key, err := cert.DecryptCertKey()
	if key == nil || err != nil {
		return fmt.Errorf("unable to decrypt private key: %w", err)
	}

	decodedKey, _ := pem.Decode(key)
	if decodedKey == nil || decodedKey.Type != "RSA PRIVATE KEY" {
		return fmt.Errorf("unable to decode private key")
	}

	parsedKey, err := x509.ParsePKCS1PrivateKey(decodedKey.Bytes)
	if err != nil {
		return fmt.Errorf("unable to parse private key: %w", err)
	}

	now := time.Now()

	newCert, err := models.GenerateCertificate(serviceName, parsedKey, now)
	if err != nil {
		return fmt.Errorf("unable to renew certificate: %w", err)
	}

	cert.CertData = newCert

	validationError, err := s.db.ValidateAndUpdate(s)
	if err != nil {
		return fmt.Errorf("unable to update certificate: %w", err)
	}

	if validationError != nil && validationError.HasAny() {
		return fmt.Errorf("saml certificate validation failed: %v", validationError)
	}

	return nil
}

func (s samlCertificatePersister) Delete(cert *models.SamlCertificate) error {
	err := s.db.Destroy(cert)
	if err != nil {
		return fmt.Errorf("failed to delete certificate: %w", err)
	}

	return nil
}
