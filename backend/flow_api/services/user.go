package services

import (
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

func UserCanDoThirdParty(cfg config.Config, identities models.Identities) bool {
	for _, identity := range identities {
		if cfg.ThirdParty.Providers.Get(identity.ProviderName).Enabled {
			return true
		}
	}

	return false
}
