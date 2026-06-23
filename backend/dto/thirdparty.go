package dto

import (
	"slices"
	"strings"

	"github.com/fatih/structs"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
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
	ID         string    `json:"id"` // the user/subject ID at the provider, ProviderUserID from models.Identity
	Provider   string    `json:"provider"`
	IdentityID uuid.UUID `json:"identity_id"` // the internal id from models.Identity
}

type Identities []Identity

func FromIdentitiesModel(identities models.Identities, cfg *config.TenantConfig) Identities {
	enabledBuiltIn := cfg.ThirdParty.Providers.GetEnabled()
	enabledCustom := cfg.ThirdParty.CustomProviders.GetEnabled()

	var result Identities
	for _, i := range identities {
		if isIdentityForDisabledProvider(i, enabledBuiltIn, enabledCustom) {
			continue
		}
		identity := FromIdentityModel(&i, cfg)
		result = append(result, *identity)
	}
	return result
}

func isIdentityForDisabledProvider(i models.Identity, enabledBuiltIn []config.ThirdPartyProvider, enabledCustom []config.CustomThirdPartyProvider) bool {
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

func FromIdentityModel(identity *models.Identity, cfg *config.TenantConfig) *Identity {
	if identity == nil {
		return nil
	}

	return &Identity{
		ID:         identity.ProviderUserID,
		Provider:   getProviderDisplayName(identity, cfg),
		IdentityID: identity.ID,
	}
}

var builtInProviderDisplayNames = func() map[string]string {
	m := make(map[string]string)
	s := structs.New(config.ThirdPartyProviders{})
	for _, f := range s.Fields() {
		m[strings.ToLower(f.Name())] = f.Name()
	}
	return m
}()

func getProviderDisplayName(identity *models.Identity, cfg *config.TenantConfig) string {
	if identity.SamlIdentity != nil {
		return identity.SamlIdentity.SamlProvider.Name
	} else if strings.HasPrefix(identity.ProviderID, "custom_") {
		key := strings.TrimPrefix(identity.ProviderID, "custom_")
		return cfg.ThirdParty.CustomProviders[key].DisplayName
	} else if name, ok := builtInProviderDisplayNames[identity.ProviderID]; ok {
		return name
	}

	return strings.TrimSpace(identity.ProviderID)
}
