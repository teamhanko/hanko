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
		Providers:               GetEnabledProviders(config.ThirdParty.Providers),
		Account:                 config.Account,
		UseEnterpriseConnection: UseEnterpriseConnection(&config.Saml),
	}
}

func GetEnabledProviders(providers config.ThirdPartyProviders) []string {
	s := structs.New(providers)
	var enabledProviders []string
	for _, field := range s.Fields() {
		v := field.Value().(config.ThirdPartyProvider)
		if v.Enabled {
			enabledProviders = append(enabledProviders, field.Name())
		}
	}
	return enabledProviders
}

func UseEnterpriseConnection(samlConfig *samlConfig.Saml) bool {
	return samlConfig != nil && samlConfig.Enabled
}
