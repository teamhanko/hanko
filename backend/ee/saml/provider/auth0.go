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

	if attributeMap.Name == "" {
		attributeMap.Name = "http://schemas.auth0.com/name"
	}

	if attributeMap.Email == "" {
		attributeMap.Name = "http://schemas.auth0.com/email"
	}

	if attributeMap.EmailVerified == "" {
		attributeMap.EmailVerified = "http://schemas.auth0.com/email_verified"
	}

	if attributeMap.NickName == "" {
		attributeMap.NickName = "http://schemas.auth0.com/nickname"
	}

	if attributeMap.Picture == "" {
		attributeMap.Picture = "http://schemas.auth0.com/picture"
	}

	if attributeMap.UpdatedAt == "" {
		attributeMap.UpdatedAt = "http://schemas.auth0.com/updated_at"
	}
}
