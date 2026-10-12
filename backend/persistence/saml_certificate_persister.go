package persistence

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

type SamlCertificatePersister interface {
	Create(cert *models.SamlCertificate) error
	GetFirst(tenantID uuid.UUID) (*models.SamlCertificate, error)
	Renew(cert *models.SamlCertificate, serviceName string, opts ...models.CertificateOption) (keyChanged bool, err error)
	Delete(cert *models.SamlCertificate) error
}

type samlCertificatePersister struct {
	db *pop.Connection
}

func NewSamlCertificatePersister(db *pop.Connection) SamlCertificatePersister {
	return &samlCertificatePersister{db: db}
}

func (s samlCertificatePersister) GetFirst(tenantID uuid.UUID) (*models.SamlCertificate, error) {
	cert := models.SamlCertificate{}

	query := s.db.Q()
	query = query.Where("tenant_id = ?", tenantID)
	err := query.First(&cert)
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
		return fmt.Errorf("saml certificate validation failed: %w", validationError)
	}

	return nil
}

func (s samlCertificatePersister) Renew(cert *models.SamlCertificate, serviceName string, opts ...models.CertificateOption) (bool, error) {
	keyChanged, err := cert.Renew(serviceName, opts...)
	if err != nil {
		return false, err
	}

	validationError, err := s.db.ValidateAndUpdate(cert)
	if err != nil {
		return false, fmt.Errorf("unable to update certificate: %w", err)
	}

	if validationError != nil && validationError.HasAny() {
		return false, fmt.Errorf("saml certificate validation failed: %v", validationError)
	}

	return keyChanged, nil
}

func (s samlCertificatePersister) Delete(cert *models.SamlCertificate) error {
	err := s.db.Destroy(cert)
	if err != nil {
		return fmt.Errorf("failed to delete certificate: %w", err)
	}

	return nil
}
