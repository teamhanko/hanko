package saml

import (
	"fmt"

	"github.com/gofrs/uuid"
	saml2 "github.com/russellhaering/gosaml2"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/persistence"
	"github.com/teamhanko/hanko/backend/v2/thirdparty"
)

type SamlProviderService interface {
	Persister() persistence.Persister
	GetProviderByDomain(tenantID uuid.UUID, tenantConfig config.TenantConfig, domain string) (*saml2.SAMLServiceProvider, *ProviderConfig, error)
	GetProviderByIssuer(tenantID uuid.UUID, tenantConfig config.TenantConfig, issuer string) (*saml2.SAMLServiceProvider, *ProviderConfig, error)
	GetAuthUrl(tenantID uuid.UUID, config config.Config, providerID uuid.UUID, redirectTo string, isFlow bool) (string, error)
}

type defaultService struct {
	persister      persistence.Persister
	runtimeBuilder *ProviderManager
}

func NewSamlProviderService(persister persistence.Persister) SamlProviderService {
	return &defaultService{
		persister:      persister,
		runtimeBuilder: NewProviderManager(persister),
	}
}

func (s *defaultService) Persister() persistence.Persister {
	return s.persister
}

// GetProviderByDomain attempts to retrieve an enabled provider for the given domain from DB and builds the runtime
// gosaml2 provider using the tenant configuration.
func (s *defaultService) GetProviderByDomain(tenantID uuid.UUID, tenantConfig config.TenantConfig, domain string) (*saml2.SAMLServiceProvider, *ProviderConfig, error) {
	// Query DB for provider by domain
	providerModel, err := s.persister.GetSamlProviderPersister().GetEnabledByDomain(tenantID, domain)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get provider by domain: %w", err)
	}
	if providerModel == nil {
		return nil, nil, fmt.Errorf("no provider found for domain: %s", domain)
	}

	// GetProvider runtime provider from cached metadata using tenant-specific config
	samlProvider, providerConfig, err := s.runtimeBuilder.GetProvider(tenantID, providerModel.ID, tenantConfig.Saml, tenantConfig.Service.Name)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build SAML provider: %w", err)
	}

	return samlProvider, providerConfig, nil
}

// GetProviderByIssuer retrieves a provider from DB by issuer and builds the runtime gosaml2 provider
func (s *defaultService) GetProviderByIssuer(tenantID uuid.UUID, tenantConfig config.TenantConfig, issuer string) (*saml2.SAMLServiceProvider, *ProviderConfig, error) {
	// Query DB for provider by entity_id (which is the issuer)
	providerModel, err := s.persister.GetSamlProviderPersister().GetByEntityID(tenantID, issuer)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get provider by issuer: %w", err)
	}
	if providerModel == nil {
		return nil, nil, fmt.Errorf("no provider found for issuer: %s", issuer)
	}

	// GetProvider runtime provider from cached metadata using tenant-specific config
	samlProvider, providerConfig, err := s.runtimeBuilder.GetProvider(tenantID, providerModel.ID, tenantConfig.Saml, tenantConfig.Service.Name)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build SAML provider: %w", err)
	}

	return samlProvider, providerConfig, nil
}

// GetAuthUrl generates a SAML authentication URL for a given provider
func (s *defaultService) GetAuthUrl(tenantID uuid.UUID, config config.Config, providerID uuid.UUID, redirectTo string, isFlow bool) (string, error) {
	if ok := config.TenantConfig.Saml.IsAllowedRedirect(redirectTo); !ok {
		return "", thirdparty.ErrorInvalidRequest(fmt.Sprintf("redirect to '%s' not allowed", redirectTo))
	}

	providerModel, err := s.persister.GetSamlProviderPersister().Get(tenantID, providerID)
	if err != nil {
		return "", fmt.Errorf("failed to get provider: %w", err)
	}
	if providerModel == nil {
		return "", fmt.Errorf("provider not found")
	}

	state, err := GenerateState(
		config,
		s.persister.GetSamlStatePersister(),
		providerModel.Domain,
		redirectTo,
		tenantID,
		GenerateStateForFlowAPI(isFlow))

	if err != nil {
		return "", thirdparty.ErrorServer("could not generate state").WithCause(err)
	}

	samlProvider, _, err := s.runtimeBuilder.GetProvider(tenantID, providerID, config.TenantConfig.Saml, config.TenantConfig.Service.Name)
	if err != nil {
		return "", thirdparty.ErrorServer("could not build SAML provider").WithCause(err)
	}

	redirectUrl, err := samlProvider.BuildAuthURL(string(state))
	if err != nil {
		return "", thirdparty.ErrorServer("could not generate auth url").WithCause(err)
	}

	return redirectUrl, nil
}
