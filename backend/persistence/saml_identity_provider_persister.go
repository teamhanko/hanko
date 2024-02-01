package persistence

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type SamlIdentityProviderPersister interface {
	List() (models.SamlIdentityProviders, error)
	Get(providerId uuid.UUID) (*models.SamlIdentityProvider, error)
	GetByDomain(domain string) (*models.SamlIdentityProvider, error)

	Create(provider *models.SamlIdentityProvider, attributeMap *models.SamlAttributeMap) error

	Update(provider *models.SamlIdentityProvider) error
	Delete(provider *models.SamlIdentityProvider) error
}

type samlIdentityProviderPersister struct {
	db *pop.Connection
}

func NewSamlIdentityProviderPersister(db *pop.Connection) SamlIdentityProviderPersister {
	return &samlIdentityProviderPersister{db: db}
}

func (s *samlIdentityProviderPersister) List() (models.SamlIdentityProviders, error) {
	list := make(models.SamlIdentityProviders, 0)
	err := s.db.Eager().All(&list)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return list, nil
}

func (s *samlIdentityProviderPersister) Get(providerId uuid.UUID) (*models.SamlIdentityProvider, error) {
	var provider models.SamlIdentityProvider
	err := s.db.Eager().Find(&provider, providerId)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &provider, nil
}

func (s *samlIdentityProviderPersister) GetByDomain(domain string) (*models.SamlIdentityProvider, error) {
	var provider models.SamlIdentityProvider
	err := s.db.Eager().Where("domain = ?", domain).First(&provider)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &provider, nil
}

func (s *samlIdentityProviderPersister) Create(provider *models.SamlIdentityProvider, attributeMap *models.SamlAttributeMap) error {
	validationError, err := s.db.ValidateAndCreate(provider)
	if err != nil {
		return err
	}

	if validationError != nil && validationError.HasAny() {
		return fmt.Errorf("saml provider validation failed: %w", validationError)
	}

	validationError, err = s.db.ValidateAndCreate(attributeMap)
	if err != nil {
		return err
	}

	if validationError != nil && validationError.HasAny() {
		return fmt.Errorf("saml provider attribute map validation failed: %w", validationError)
	}

	return nil
}

func (s *samlIdentityProviderPersister) Update(provider *models.SamlIdentityProvider) error {
	validationError, err := s.db.ValidateAndUpdate(provider)
	if err != nil {
		return err
	}

	if validationError != nil && validationError.HasAny() {
		return fmt.Errorf("saml provider validation failed: %w", validationError)
	}

	validationError, err = s.db.ValidateAndUpdate(&provider.AttributeMap)
	if err != nil {
		return err
	}

	if validationError != nil && validationError.HasAny() {
		return fmt.Errorf("saml provider attribute map validation failed: %w", validationError)
	}

	return nil
}

func (s *samlIdentityProviderPersister) Delete(provider *models.SamlIdentityProvider) error {
	err := s.db.Destroy(provider)

	if err != nil {
		return fmt.Errorf("failed to delete saml provider: %w", err)
	}

	return nil
}
