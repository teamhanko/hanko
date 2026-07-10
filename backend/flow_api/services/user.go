package services

import (
	"regexp"

	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
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
	if cfg.Saml.Enabled {
		for _, identity := range identities {
			return identity.SamlIdentity != nil && identity.SamlIdentity.SamlProvider.Enabled
		}
	}

	return false
}

func ValidateUsername(name string) bool {
	re := regexp.MustCompile(`^\w+$`)
	return re.MatchString(name)
}
