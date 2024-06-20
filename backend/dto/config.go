package dto

import (
	"github.com/fatih/structs"
	"github.com/teamhanko/hanko/backend/config"
	samlConfig "github.com/teamhanko/hanko/backend/ee/saml/config"
)

// PublicConfig is the part of the configuration that will be shared with the frontend
type PublicConfig struct {
	Password                Password `json:"password"`
	Emails                  Emails   `json:"emails"`
	Providers               []string `json:"providers"`
	Account                 Account  `json:"account"`
	UseEnterpriseConnection bool     `json:"use_enterprise"`
}

type Password struct {
	Enabled   bool `json:"enabled"`
	MinLength int  `json:"min_password_length"`
}

type Emails struct {
	RequireVerification bool `json:"require_verification"`
	MaxNumOfAddresses   int  `json:"max_num_of_addresses"`
}

type Account struct {
	AllowDeletion bool `json:"allow_deletion"`
	AllowSignup   bool `json:"allow_signup"`
}

// FromConfig Returns a PublicConfig from the Application configuration
func FromConfig(cfg config.Config) PublicConfig {
	return PublicConfig{
		Password: Password{
			Enabled:   cfg.Password.Enabled,
			MinLength: cfg.Password.MinLength,
		},
		Emails: Emails{
			RequireVerification: cfg.Email.RequireVerification,
			MaxNumOfAddresses:   cfg.Email.Limit,
		},
		Providers: GetEnabledProviders(cfg.ThirdParty.Providers),
		Account: Account{
			AllowDeletion: cfg.Account.AllowDeletion,
			AllowSignup:   cfg.Account.AllowSignup,
		},
		UseEnterpriseConnection: UseEnterpriseConnection(&cfg.Saml),
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
