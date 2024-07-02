package saml

import (
	"fmt"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/ee/saml/provider"
	samlUtils "github.com/teamhanko/hanko/backend/ee/saml/utils"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/thirdparty"
	"net/url"
)

type Service interface {
	Config() *config.Config
	Persister() persistence.Persister
	Providers() []provider.ServiceProvider
	GetProviderByDomain(domain string) (provider.ServiceProvider, error)
	GetAuthUrl(provider provider.ServiceProvider, redirectTo string, isFlow bool) (string, error)
}

type defaultService struct {
	config    *config.Config
	persister persistence.Persister
	providers []provider.ServiceProvider
}

func NewSamlService(cfg *config.Config, persister persistence.Persister) Service {
	providers := make([]provider.ServiceProvider, 0)
	for _, idpConfig := range cfg.Saml.IdentityProviders {
		if idpConfig.Enabled {
			hostName := ""
			hostName, err := parseProviderFromMetadataUrl(idpConfig.MetadataUrl)
			if err != nil {
				fmt.Printf("failed to parse provider '%s' from metadata url: %v\n", idpConfig.Name, err)
				continue
			}

			newProvider, err := provider.GetProvider(hostName, cfg, idpConfig, persister.GetSamlCertificatePersister())
			if err != nil {
				fmt.Printf("failed to initialize provider '%s': %v\n", idpConfig.Name, err)
				continue
			}

			providers = append(providers, newProvider)
		}
	}

	return &defaultService{
		config:    cfg,
		persister: persister,
		providers: providers,
	}
}

func parseProviderFromMetadataUrl(idpUrlString string) (string, error) {
	idpUrl, err := url.Parse(idpUrlString)
	if err != nil {
		return "", err
	}

	return idpUrl.Host, nil
}

func (s *defaultService) Config() *config.Config {
	return s.config
}

func (s *defaultService) Persister() persistence.Persister {
	return s.persister
}

func (s *defaultService) Providers() []provider.ServiceProvider {
	return s.providers
}

func (s *defaultService) GetProviderByDomain(domain string) (provider.ServiceProvider, error) {
	for _, availableProvider := range s.providers {
		if availableProvider.GetDomain() == domain {
			return availableProvider, nil
		}
	}

	return nil, fmt.Errorf("unknown provider for domain %s", domain)
}

func (s *defaultService) GetAuthUrl(provider provider.ServiceProvider, redirectTo string, isFlow bool) (string, error) {
	if ok := samlUtils.IsAllowedRedirect(s.config.Saml, redirectTo); !ok {
		return "", thirdparty.ErrorInvalidRequest(fmt.Sprintf("redirect to '%s' not allowed", redirectTo))
	}

	state, err := GenerateState(
		s.config,
		s.persister.GetSamlStatePersister(),
		provider.GetDomain(),
		redirectTo,
		GenerateStateForFlowAPI(isFlow))

	if err != nil {
		return "", thirdparty.ErrorServer("could not generate state").WithCause(err)
	}

	redirectUrl, err := provider.GetService().BuildAuthURL(string(state))
	if err != nil {
		return "", thirdparty.ErrorServer("could not generate auth url").WithCause(err)
	}

	return redirectUrl, nil
}
