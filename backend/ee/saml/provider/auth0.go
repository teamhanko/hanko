package provider

import (
	"github.com/teamhanko/hanko/backend/config"
	samlConfig "github.com/teamhanko/hanko/backend/ee/saml/config"
	"github.com/teamhanko/hanko/backend/persistence"
)

type Auth0Provider struct {
	*BaseSamlProvider
}

func NewAuth0ServiceProvider(config *config.Config, idpConfig samlConfig.IdentityProvider, persister persistence.SamlCertificatePersister) (ServiceProvider, error) {
	serviceProvider, err := NewBaseSamlProvider(config, idpConfig, persister)
	if err != nil {
		return nil, err
	}

	provider := &Auth0Provider{
		serviceProvider.(*BaseSamlProvider),
	}
	provider.UseDefaultAttributesIfEmpty()

	return provider, nil
}

func (sp *Auth0Provider) UseDefaultAttributesIfEmpty() {
	attributeMap := &sp.Config.AttributeMap

	if attributeMap.EmailVerified == "" {
		attributeMap.EmailVerified = "email_verified"
	}

	if attributeMap.NickName == "" {
		attributeMap.NickName = "nickname"
	}

	if attributeMap.Picture == "" {
		attributeMap.Picture = "picture"
	}

	if attributeMap.UpdatedAt == "" {
		attributeMap.UpdatedAt = "updated_at"
	}
}
