package saml

import (
	"fmt"
	"log"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/persistence"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
)

// SyncProviderConfigToDatabase syncs SAML providers from config file to database
// This is called at application startup to maintain backward compatibility
// for single-tenant deployments using config-based provider definitions
func SyncProviderConfigToDatabase(cfg *config.Config, persister persistence.Persister) error {
	// Only attempt sync if Mutlitenancy is not enabled and if SAML is enabled
	if cfg.MultiTenancy.Enabled {
		log.Println("[SAML Migration] Multitenancy enabled, skipping config sync")
		return nil
	}

	if !cfg.Saml.Enabled {
		log.Println("[SAML Migration] SAML is disabled, skipping config sync")
		return nil
	}

	// Check if there are any providers in config
	if len(cfg.Saml.IdentityProviders) == 0 {
		log.Println("[SAML Migration] No identity providers in config, skipping sync")
		return nil
	}

	// Use default tenant ID for single-tenant mode
	tenantID, err := uuid.FromString(config.DefaultTenantID)
	if err != nil {
		return fmt.Errorf("invalid default tenant ID: %w", err)
	}

	// Ensure default tenant exists
	err = ensureDefaultTenantExists(persister, tenantID)
	if err != nil {
		return fmt.Errorf("failed to ensure default tenant exists: %w", err)
	}

	err = ensureSamlCertificateExists(persister, tenantID, cfg.Service.Name)

	providerSvc := NewSamlProviderManagementService(persister)

	successCount := 0
	failCount := 0

	log.Printf("[SAML Migration] Syncing %d identity providers from config to database...", len(cfg.Saml.IdentityProviders))

	for _, idpConfig := range cfg.Saml.IdentityProviders {
		// Skip disabled providers
		if !idpConfig.Enabled {
			log.Printf("[SAML Migration] Skipping disabled provider: %s (domain: %s)", idpConfig.Name, idpConfig.Domain)
			continue
		}

		log.Printf("[SAML Migration] Syncing provider: %s (domain: %s, metadata: %s)",
			idpConfig.Name, idpConfig.Domain, idpConfig.MetadataUrl)

		err := providerSvc.CreateFromConfig(tenantID, idpConfig)
		if err != nil {
			// Log error but don't fail startup
			log.Printf("[SAML Migration] WARNING: Failed to sync provider '%s': %v", idpConfig.Name, err)
			failCount++
			continue
		}

		successCount++
		log.Printf("[SAML Migration] Successfully synced provider: %s", idpConfig.Name)
	}

	if failCount > 0 {
		log.Printf("[SAML Migration] Completed with warnings: %d succeeded, %d failed", successCount, failCount)
	} else {
		log.Printf("[SAML Migration] Successfully synced %d providers from config", successCount)
	}

	return nil
}

// ensureDefaultTenantExists creates the default tenant if it doesn't exist
func ensureDefaultTenantExists(persister persistence.Persister, tenantID uuid.UUID) error {
	tenant, err := persister.GetTenantPersister().Get(tenantID)
	if err != nil {
		return fmt.Errorf("failed to check if default tenant exists: %w", err)
	}

	// Tenant already exists
	if tenant != nil {
		return nil
	}

	// Note: In single-tenant mode, the default tenant should be created elsewhere
	// (e.g., during initial database setup or first-time configuration)
	// This function just verifies it exists
	log.Printf("[SAML Migration] WARNING: Default tenant does not exist (ID: %s)", tenantID)
	log.Printf("[SAML Migration] SAML providers will not be synced until default tenant is created")

	return fmt.Errorf("default tenant does not exist")
}

func ensureSamlCertificateExists(persister persistence.Persister, tenantID uuid.UUID, serviceName string) error {
	return persister.Transaction(func(tx *pop.Connection) error {
		certPersister := persister.GetSamlCertificatePersisterWithConnection(tx)

		cert, err := certPersister.GetFirst(tenantID)
		if err != nil {
			return fmt.Errorf("failed to fetch SAML certificate: %w", err)
		}

		if cert == nil {
			log.Printf("[SAML Migration] No SAML certificate for tenant '%s' exists, creating certificate ...", tenantID.String())

			cert, err = models.NewSamlCertificate(serviceName)
			if err != nil {
				return fmt.Errorf("unable to create SAML certificate: %w", err)
			}

			cert.TenantID = tenantID

			err = certPersister.Create(cert)
			if err != nil {
				return fmt.Errorf("unable to persist SAML certificate: %w", err)
			}
			log.Printf("[SAML Migration] Successfully created SAML certificate for tenant '%s", tenantID.String())
		}

		return nil
	})
}
