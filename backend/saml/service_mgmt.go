package saml

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	samlConfig "github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/persistence"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

// SamlProviderManagementService handles SAML provider lifecycle operations
type SamlProviderManagementService struct {
	persister       persistence.Persister
	metadataService *SamlMetadataService
}

func NewSamlProviderManagementService(persister persistence.Persister) *SamlProviderManagementService {
	return &SamlProviderManagementService{
		persister:       persister,
		metadataService: NewSamlMetadataService(persister),
	}
}

// CreateFromConfig creates a provider from config (used for auto-migration)
// This is called at startup to sync config-based providers into the database
func (s *SamlProviderManagementService) CreateFromConfig(tenantID uuid.UUID, idpConfig samlConfig.IdentityProvider) error {
	// Fetch and parse metadata (outside transaction since it's an external HTTP call)
	parsedMetadata, err := s.metadataService.FetchAndParse(idpConfig.MetadataUrl)
	if err != nil {
		return fmt.Errorf("failed to fetch metadata for provider '%s': %w", idpConfig.Name, err)
	}

	// Serialize attribute map to JSON
	attributeMapJSON, err := json.Marshal(idpConfig.AttributeMap)
	if err != nil {
		return fmt.Errorf("failed to marshal attribute map: %w", err)
	}

	return s.persister.Transaction(func(tx *pop.Connection) error {
		providerPersister := s.persister.GetSamlProviderPersisterWithConnection(tx)

		existing, err := providerPersister.GetByDomain(tenantID, idpConfig.Domain)
		if err != nil {
			return fmt.Errorf("failed to check existing provider: %w", err)
		}

		if existing != nil {
			existing.Name = idpConfig.Name
			existing.MetadataURL = idpConfig.MetadataUrl
			existing.EntityID = parsedMetadata.EntityID
			existing.Enabled = idpConfig.Enabled
			existing.SkipEmailVerification = idpConfig.SkipEmailVerification
			existing.AttributeMap = string(attributeMapJSON)

			err = providerPersister.Update(*existing)
			if err != nil {
				return fmt.Errorf("failed to update provider: %w", err)
			}

			return s.storeMetadataInTransaction(tx, tenantID, existing.ID, parsedMetadata)
		}

		provider := models.SamlProvider{
			ID:                    uuid.Must(uuid.NewV4()),
			TenantID:              tenantID,
			Name:                  idpConfig.Name,
			EntityID:              parsedMetadata.EntityID,
			MetadataURL:           idpConfig.MetadataUrl,
			Domain:                idpConfig.Domain,
			Enabled:               idpConfig.Enabled,
			SkipEmailVerification: idpConfig.SkipEmailVerification,
			AttributeMap:          string(attributeMapJSON),
		}

		err = providerPersister.Create(provider)
		if err != nil {
			return fmt.Errorf("failed to create provider: %w", err)
		}

		// Store metadata cache within same transaction
		return s.storeMetadataInTransaction(tx, tenantID, provider.ID, parsedMetadata)
	})
}

// Create creates a new SAML provider (for API use)
func (s *SamlProviderManagementService) Create(tenantID uuid.UUID, name, metadataURL, domain string, enabled, skipEmailVerification bool, attributeMap samlConfig.AttributeMap) (*models.SamlProvider, error) {
	// Fetch and parse metadata (outside transaction since it's an external HTTP call)
	parsedMetadata, err := s.metadataService.FetchAndParse(metadataURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metadata: %w", err)
	}

	// Serialize attribute map
	attributeMapJSON, err := json.Marshal(attributeMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal attribute map: %w", err)
	}

	var provider models.SamlProvider
	err = s.persister.Transaction(func(tx *pop.Connection) error {
		providerPersister := s.persister.GetSamlProviderPersisterWithConnection(tx)

		existingByDomain, err := providerPersister.GetByDomain(tenantID, domain)
		if err != nil {
			return err
		}
		if existingByDomain != nil {
			return fmt.Errorf("provider with domain '%s' already exists", domain)
		}

		existingByEntityID, err := providerPersister.GetByEntityID(tenantID, parsedMetadata.EntityID)
		if err != nil {
			return err
		}
		if existingByEntityID != nil {
			return fmt.Errorf("provider with entity_id '%s' already exists", parsedMetadata.EntityID)
		}

		// Create provider
		provider = models.SamlProvider{
			ID:                    uuid.Must(uuid.NewV4()),
			TenantID:              tenantID,
			Name:                  name,
			EntityID:              parsedMetadata.EntityID,
			MetadataURL:           metadataURL,
			Domain:                domain,
			Enabled:               enabled,
			SkipEmailVerification: skipEmailVerification,
			AttributeMap:          string(attributeMapJSON),
		}

		err = providerPersister.Create(provider)
		if err != nil {
			return fmt.Errorf("failed to create provider: %w", err)
		}

		// Store metadata within same transaction
		return s.storeMetadataInTransaction(tx, tenantID, provider.ID, parsedMetadata)
	})

	if err != nil {
		return nil, err
	}

	return &provider, nil
}

// Update updates an existing SAML provider
func (s *SamlProviderManagementService) Update(tenantID uuid.UUID, providerID uuid.UUID, name, metadataURL, domain string, enabled, skipEmailVerification bool, attributeMap samlConfig.AttributeMap) error {
	samlProvider, err := s.persister.GetSamlProviderPersister().Get(tenantID, providerID)
	if err != nil {
		return err
	}
	if samlProvider == nil {
		return fmt.Errorf("provider not found")
	}

	// Fetch new metadata if URL changed (outside transaction since it's an external HTTP call)
	var parsedMetadata *ParsedMetadata
	if metadataURL != samlProvider.MetadataURL {
		parsedMetadata, err = s.metadataService.FetchAndParse(metadataURL)
		if err != nil {
			return fmt.Errorf("failed to fetch metadata: %w", err)
		}
	}

	// Serialize attribute map
	attributeMapJSON, err := json.Marshal(attributeMap)
	if err != nil {
		return fmt.Errorf("failed to marshal attribute map: %w", err)
	}

	// Wrap all database operations in a transaction
	return s.persister.Transaction(func(tx *pop.Connection) error {
		providerPersister := s.persister.GetSamlProviderPersisterWithConnection(tx)

		// Re-fetch provider within transaction to ensure consistency
		provider, err := providerPersister.Get(tenantID, providerID)
		if err != nil {
			return err
		}
		if provider == nil {
			return fmt.Errorf("provider not found")
		}

		if parsedMetadata != nil {
			provider.EntityID = parsedMetadata.EntityID
			provider.MetadataURL = metadataURL

			err = s.storeMetadataInTransaction(tx, tenantID, providerID, parsedMetadata)
			if err != nil {
				return fmt.Errorf("failed to update metadata: %w", err)
			}
		}

		// Update provider fields
		provider.Name = name
		provider.Domain = domain
		provider.Enabled = enabled
		provider.SkipEmailVerification = skipEmailVerification
		provider.AttributeMap = string(attributeMapJSON)

		return providerPersister.Update(*provider)
	})
}

// Get retrieves a provider by ID
func (s *SamlProviderManagementService) Get(tenantID uuid.UUID, providerID uuid.UUID) (*models.SamlProvider, error) {
	return s.persister.GetSamlProviderPersister().Get(tenantID, providerID)
}

// GetByDomain retrieves a provider by domain
func (s *SamlProviderManagementService) GetByDomain(tenantID uuid.UUID, domain string) (*models.SamlProvider, error) {
	return s.persister.GetSamlProviderPersister().GetByDomain(tenantID, domain)
}

// List retrieves all providers for a tenant
func (s *SamlProviderManagementService) List(tenantID uuid.UUID) ([]models.SamlProvider, error) {
	return s.persister.GetSamlProviderPersister().List(tenantID)
}

// Delete deletes a provider
func (s *SamlProviderManagementService) Delete(tenantID uuid.UUID, providerID uuid.UUID) error {
	return s.persister.Transaction(func(tx *pop.Connection) error {
		providerPersister := s.persister.GetSamlProviderPersisterWithConnection(tx)

		provider, err := providerPersister.Get(tenantID, providerID)
		if err != nil {
			return err
		}
		if provider == nil {
			return fmt.Errorf("provider not found")
		}

		// Delete Identities for this provider, since they become useless and no DB level cascade is available
		identityPersister := s.persister.GetIdentityPersisterWithConnection(tx)
		identities, err := identityPersister.GetAllByDomain(tenantID, provider.Domain)
		if err != nil {
			return fmt.Errorf("failed to get identities for provider: %w", err)
		}

		if err := identityPersister.DeleteAll(identities); err != nil {
			return fmt.Errorf("failed to delete identities for provider: %w", err)
		}

		// Delete provider (metadata will be automatically deleted via cascade foreign key)
		return providerPersister.Delete(*provider)
	})
}

// storeMetadataInTransaction stores metadata within an existing transaction
func (s *SamlProviderManagementService) storeMetadataInTransaction(tx *pop.Connection, tenantID uuid.UUID, providerID uuid.UUID, metadata *ParsedMetadata) error {
	certsJSON, err := json.Marshal(metadata.CertificatesPEM)
	if err != nil {
		return fmt.Errorf("failed to marshal certificates: %w", err)
	}

	metadataModel := models.SamlIDPMetadata{
		ID:              uuid.Must(uuid.NewV4()),
		TenantID:        tenantID,
		ProviderID:      providerID,
		RawMetadataXML:  metadata.RawXML,
		Issuer:          metadata.Issuer,
		SSOURL:          metadata.SSOURL,
		CertificatesPEM: string(certsJSON),
		LastFetchedAt:   time.Now(),
	}

	metadataPersister := s.persister.GetSamlIDPMetadataPersisterWithConnection(tx)

	existing, err := metadataPersister.Get(tenantID, providerID)
	if err != nil {
		return fmt.Errorf("failed to check existing metadata: %w", err)
	}

	if existing != nil {
		metadataModel.ID = existing.ID
		metadataModel.CreatedAt = existing.CreatedAt
		return metadataPersister.Update(metadataModel)
	}

	return metadataPersister.Create(metadataModel)
}
