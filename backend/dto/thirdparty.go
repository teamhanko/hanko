package dto

import (
	"github.com/fatih/structs"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"strings"
)

type ThirdPartyAuthCallback struct {
	AuthCode         string `query:"code"`
	State            string `query:"state" validate:"required"`
	Error            string `query:"error"`
	ErrorDescription string `query:"error_description"`
}

func (cb ThirdPartyAuthCallback) HasError() bool {
	return cb.Error != ""
}

type ThirdPartyAuthRequest struct {
	Provider   string `query:"provider" validate:"required"`
	RedirectTo string `query:"redirect_to" validate:"required,url"`
}

type Identity struct {
	ID       string `json:"id"`
	Provider string `json:"provider"`
}

type Identities []Identity

func FromIdentitiesModel(identities models.Identities, cfg *config.Config) Identities {
	var result Identities
	for _, i := range identities {
		identity := FromIdentityModel(&i, cfg)
		result = append(result, *identity)
	}
	return result
}

func FromIdentityModel(identity *models.Identity, cfg *config.Config) *Identity {
	if identity == nil {
		return nil
	}

	return &Identity{
		ID:       identity.ProviderID,
		Provider: getProviderDisplayName(identity, cfg),
	}
}

func getProviderDisplayName(identity *models.Identity, cfg *config.Config) string {
	if strings.HasPrefix(identity.ProviderName, "custom_") {
		providerNameWithoutPrefix := strings.TrimPrefix(identity.ProviderName, "custom_")
		return cfg.ThirdParty.CustomProviders[providerNameWithoutPrefix].DisplayName
	} else {
		s := structs.New(config.ThirdPartyProviders{})
		for _, field := range s.Fields() {
			if strings.ToLower(field.Name()) == strings.ToLower(identity.ProviderName) {
				return field.Name()
			}
		}
	}

	return strings.TrimSpace(identity.ProviderName)
}
