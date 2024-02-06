package dto

import (
	"github.com/fatih/structs"
	"github.com/teamhanko/hanko/backend/config"
	samlConfig "github.com/teamhanko/hanko/backend/ee/saml/config"
)

// PublicConfig is the part of the configuration that will be shared with the frontend
type PublicConfig struct {
	Password                config.Password `json:"password"`
	Emails                  config.Emails   `json:"emails"`
	Providers               []string        `json:"providers"`
	Account                 config.Account  `json:"account"`
	UseEnterpriseConnection bool            `json:"use_enterprise"`
}

// FromConfig Returns a PublicConfig from the Application configuration
func FromConfig(config config.Config) PublicConfig {
	return PublicConfig{
		Password:                config.Password,
		Emails:                  config.Emails,
		Providers:               GetEnabledProviders(config.ThirdParty),
		Account:                 config.Account,
		UseEnterpriseConnection: UseEnterpriseConnection(&config.Saml),
	}
}

func GetEnabledProviders(thirdParty config.ThirdParty) []string {
	// TODO: a provider should return an slug and the display name.
	// It looks like the name, aka what is displayed is also used to request and auth
	providers := thirdParty.Providers
	s := structs.New(providers)
	var enabledProviders []string
	for _, field := range s.Fields() {
		v := field.Value().(config.ThirdPartyProvider)
		if v.Enabled && !v.Hidden {
			displayName := field.Name()
			/*
				// fails because the display name is uses as the lookup key.
				// need to send the client both.
				if v.DisplayName != "" {
					displayName = v.DisplayName
				}
			*/
			enabledProviders = append(enabledProviders, displayName)
		}
	}
	if thirdParty.GenericOIDCProviders != nil {
		for k, v := range thirdParty.GenericOIDCProviders {
			if v.Enabled && !v.Hidden {
				displayName := k
				/*
					// fails because the display name is uses as the lookup key.
					// need to send the client both.
					if v.DisplayName != "" {
						displayName = v.DisplayName
					}
				*/
				enabledProviders = append(enabledProviders, displayName)
			}
		}
	}

	return enabledProviders
}

func UseEnterpriseConnection(samlConfig *samlConfig.Saml) bool {
	hasProvider := false

	if samlConfig != nil && samlConfig.Enabled {
		for _, availableProvider := range samlConfig.IdentityProviders {
			if availableProvider.Enabled {
				hasProvider = true
			}
		}
	}

	return hasProvider

}
