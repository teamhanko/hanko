package local_db

import (
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/persistence"
)

func SyncSecretKeys(cfg *config.Config, persister persistence.Storage) error {
	if !cfg.ApplicationConfig.MultiTenancy.Enabled && cfg.TenantConfig.Secrets.KeyManagement.Type == config.KEY_MANAGEMENT_STORE_LOCAL {
		jwkPersister := persister.GetJwkPersister()
		jwkManager, err := NewDefaultManager(cfg.SecretKeys, jwkPersister)
		if err != nil {
			return fmt.Errorf("failed to create JWK manager: %w", err)
		}
		// for every key we should check if a jwk with index exists and create one if not.
		for i := range cfg.SecretKeys {
			j, err := jwkPersister.Get(i+1, uuid.FromStringOrNil(config.DefaultTenantID))
			if j == nil && err == nil {
				_, err := jwkManager.GenerateKey(uuid.FromStringOrNil(config.DefaultTenantID))
				if err != nil {
					return fmt.Errorf("failed to generate JWK: %w", err)
				}
			} else if err != nil {
				return fmt.Errorf("failed to retrieve JWK: %w", err)
			}
		}
	}

	return nil
}
