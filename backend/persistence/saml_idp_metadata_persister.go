package persistence

import (
	"fmt"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

type SamlIDPMetadataPersister interface {
	Create(metadata models.SamlIDPMetadata) error
	Update(metadata models.SamlIDPMetadata) error
	Get(tenantID uuid.UUID, providerID uuid.UUID) (*models.SamlIDPMetadata, error)
	Delete(metadata models.SamlIDPMetadata) error
}

type samlIDPMetadataPersister struct {
	db *pop.Connection
}

func NewSamlIDPMetadataPersister(db *pop.Connection) SamlIDPMetadataPersister {
	return &samlIDPMetadataPersister{db: db}
}

func (p *samlIDPMetadataPersister) Create(metadata models.SamlIDPMetadata) error {
	vErr, err := p.db.ValidateAndCreate(&metadata)
	if err != nil {
		return fmt.Errorf("failed to store saml idp metadata: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("saml idp metadata object validation failed: %w", vErr)
	}

	return nil
}

func (p *samlIDPMetadataPersister) Update(metadata models.SamlIDPMetadata) error {
	vErr, err := p.db.ValidateAndUpdate(&metadata)
	if err != nil {
		return fmt.Errorf("failed to update saml idp metadata: %w", err)
	}

	if vErr != nil && vErr.HasAny() {
		return fmt.Errorf("saml idp metadata object validation failed: %w", vErr)
	}

	return nil
}

func (p *samlIDPMetadataPersister) Get(tenantID uuid.UUID, providerID uuid.UUID) (*models.SamlIDPMetadata, error) {
	metadata := &models.SamlIDPMetadata{}
	err := p.db.Where("tenant_id = ? AND provider_id = ?", tenantID, providerID).First(metadata)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get saml idp metadata: %w", err)
	}

	return metadata, nil
}

func (p *samlIDPMetadataPersister) Delete(metadata models.SamlIDPMetadata) error {
	err := p.db.Destroy(&metadata)
	if err != nil {
		return fmt.Errorf("failed to delete saml idp metadata: %w", err)
	}

	return nil
}
