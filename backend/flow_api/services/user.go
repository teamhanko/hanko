package services

import (
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

func UserCanDoThirdParty(cfg config.Config, identities models.Identities) bool {
	for _, identity := range identities {
		if provider := cfg.ThirdParty.Providers.Get(identity.ProviderName); provider != nil {
			return provider.Enabled
		}
	}

	return false
}

func UserCanDoSaml(cfg config.Config, identities models.Identities) bool {
	for _, identity := range identities {
		if provider := cfg.Saml.GetProviderByDomain(identity.ProviderName); provider != nil {
			return cfg.Saml.Enabled && provider.Enabled
		}
	}

	return false
}
