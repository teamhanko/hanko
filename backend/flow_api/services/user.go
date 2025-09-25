package services

import (
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
	"regexp"
)

func UserCanDoThirdParty(cfg config.Config, identities models.Identities) bool {
	for _, identity := range identities {
		if provider := cfg.ThirdParty.Providers.Get(identity.ProviderID); provider != nil {
			return provider.Enabled
		}
	}

	return false
}

func UserCanDoSaml(cfg config.Config, identities models.Identities) bool {
	for _, identity := range identities {
		if provider := cfg.Saml.GetProviderByDomain(identity.ProviderID); provider != nil {
			return cfg.Saml.Enabled && provider.Enabled
		}
	}

	return false
}

func ValidateUsername(name string) bool {
	re := regexp.MustCompile(`^\w+$`)
	return re.MatchString(name)
}
