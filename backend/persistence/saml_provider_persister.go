package persistence

import (
	"fmt"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
)

type SamlProviderPersister interface {
	Create(provider models.SamlProvider) error
	Update(provider models.SamlProvider) error
	Get(tenantID uuid.UUID, providerID uuid.UUID) (*models.SamlProvider, error)
	GetByEntityID(tenantID uuid.UUID, entityID string) (*models.SamlProvider, error)
	GetByDomain(tenantID uuid.UUID, domain string) (*models.SamlProvider, error)
	GetEnabledByDomain(tenantID uuid.UUID, domain string) (*models.SamlProvider, error)
	List(tenantID uuid.UUID) ([]models.SamlProvider, error)
	Delete(provider models.SamlProvider) error
}

type samlProviderPersister struct {
	db *pop.Connection
}

func NewSamlProviderPersister(db *pop.Connection) SamlProviderPersister {
	return &samlProviderPersister{db: db}
}

func (p *samlProviderPersister) Create(provider models.SamlProvider) error {
	vErr, err := p.db.ValidateAndCreate(&provider)
	if err != nil {
		return fmt.Errorf("failed to store SAML provider: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("SAML provider object validation failed: %w", vErr)
	}

	return nil
}

func (p *samlProviderPersister) Update(provider models.SamlProvider) error {
	vErr, err := p.db.ValidateAndUpdate(&provider)
	if err != nil {
		return fmt.Errorf("failed to update SAML provider: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("SAML provider object validation failed: %w", vErr)
	}

	return nil
}

func (p *samlProviderPersister) Get(tenantID uuid.UUID, providerID uuid.UUID) (*models.SamlProvider, error) {
	provider := &models.SamlProvider{}
	err := p.db.Where("tenant_id = ? AND id = ?", tenantID, providerID).First(provider)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get SAML provider: %w", err)
	}

	return provider, nil
}

func (p *samlProviderPersister) GetByEntityID(tenantID uuid.UUID, entityID string) (*models.SamlProvider, error) {
	provider := &models.SamlProvider{}
	err := p.db.Where("tenant_id = ? AND entity_id = ?", tenantID, entityID).First(provider)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get SAML provider by entity_id: %w", err)
	}

	return provider, nil
}

func (p *samlProviderPersister) GetByDomain(tenantID uuid.UUID, domain string) (*models.SamlProvider, error) {
	provider := &models.SamlProvider{}
	err := p.db.Where("tenant_id = ? AND domain = ?", tenantID, domain).First(provider)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get SAML provider by domain: %w", err)
	}

	return provider, nil
}

func (p *samlProviderPersister) GetEnabledByDomain(tenantID uuid.UUID, domain string) (*models.SamlProvider, error) {
	provider := &models.SamlProvider{}
	err := p.db.Where("tenant_id = ? AND domain = ? AND enabled = ?", tenantID, domain, true).First(provider)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get SAML provider by domain: %w", err)
	}

	return provider, nil
}

func (p *samlProviderPersister) List(tenantID uuid.UUID) ([]models.SamlProvider, error) {
	providers := []models.SamlProvider{}
	err := p.db.Where("tenant_id = ?", tenantID).All(&providers)
	if err != nil {
		return nil, fmt.Errorf("failed to list SAML providers: %w", err)
	}

	return providers, nil
}

func (p *samlProviderPersister) Delete(provider models.SamlProvider) error {
	err := p.db.Destroy(&provider)
	if err != nil {
		return fmt.Errorf("failed to delete SAML provider: %w", err)
	}

	return nil
}
