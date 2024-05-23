package provider

import (
	"github.com/teamhanko/hanko/backend/config"
	samlConfig "github.com/teamhanko/hanko/backend/ee/saml/config"
	"github.com/teamhanko/hanko/backend/persistence"
	"strings"
)

type Auth0Provider struct {
	*BaseSamlProvider
}

func NewAuth0ServiceProvider(config *config.Config, idpConfig samlConfig.IdentityProvider, persister persistence.SamlCertificatePersister) (ServiceProvider, error) {
	serviceProvider, err := NewBaseSamlProvider(config, idpConfig, persister, false)
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

	if strings.TrimSpace(attributeMap.Name) == "" {
		attributeMap.Name = "http://schemas.auth0.com/name"
	}

	if strings.TrimSpace(attributeMap.Email) == "" {
		attributeMap.Email = "http://schemas.auth0.com/email"
	}

	if strings.TrimSpace(attributeMap.EmailVerified) == "" {
		attributeMap.EmailVerified = "http://schemas.auth0.com/email_verified"
	}

	if strings.TrimSpace(attributeMap.NickName) == "" {
		attributeMap.NickName = "http://schemas.auth0.com/nickname"
	}

	if strings.TrimSpace(attributeMap.Picture) == "" {
		attributeMap.Picture = "http://schemas.auth0.com/picture"
	}

	if strings.TrimSpace(attributeMap.UpdatedAt) == "" {
		attributeMap.UpdatedAt = "http://schemas.auth0.com/updated_at"
	}
}
