package utils

import (
	"slices"
	"strings"

	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
)

func IsIdentityForDisabledProvider(i models.Identity, enabledBuiltIn []config.ThirdPartyProvider, enabledCustom []config.CustomThirdPartyProvider) bool {
	if i.SamlIdentity != nil {
		return !i.SamlIdentity.SamlProvider.Enabled
	}

	if strings.HasPrefix(i.ProviderID, "custom_") {
		return !slices.ContainsFunc(enabledCustom, func(p config.CustomThirdPartyProvider) bool {
			return p.ID == i.ProviderID
		})
	}

	return !slices.ContainsFunc(enabledBuiltIn, func(p config.ThirdPartyProvider) bool {
		return p.ID == i.ProviderID
	})
}
