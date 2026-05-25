package saml

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strings"

	"github.com/gofrs/uuid"
	saml2 "github.com/russellhaering/gosaml2"
	dsig "github.com/russellhaering/goxmldsig"
	samlConfig "github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/persistence"
)

// ProviderManager constructs ephemeral gosaml2.SAMLServiceProvider instances from cached metadata
type ProviderManager struct {
	persister       persistence.Persister
	metadataService *SamlMetadataService
}

// ProviderConfig contains the configuration and attribute map for a provider
type ProviderConfig struct {
	AttributeMap          samlConfig.AttributeMap
	SkipEmailVerification bool
	Domain                string
}

func NewProviderManager(persister persistence.Persister) *ProviderManager {
	return &ProviderManager{
		persister:       persister,
		metadataService: NewSamlMetadataService(persister),
	}
}

// GetProvider creates an ephemeral gosaml2.SAMLServiceProvider from cached metadata
func (b *ProviderManager) GetProvider(
	tenantID uuid.UUID,
	providerID uuid.UUID,
	tenantSettings samlConfig.Saml,
	serviceName string,
) (*saml2.SAMLServiceProvider, *ProviderConfig, error) {
	// Get provider from DB
	provider, err := b.persister.GetSamlProviderPersister().Get(tenantID, providerID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get provider: %w", err)
	}
	if provider == nil {
		return nil, nil, fmt.Errorf("provider not found")
	}

	// Get cached metadata
	cachedMetadata, err := b.metadataService.Get(tenantID, providerID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get metadata: %w", err)
	}
	if cachedMetadata == nil {
		return nil, nil, fmt.Errorf("metadata not found for provider")
	}

	// Load SP certificate
	spCertStore, err := b.loadServiceProviderCertificate(tenantID, serviceName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load SP certificate: %w", err)
	}

	// GetProvider IdP certificate store from cached PEM certificates
	idpCertStore, err := b.buildIdPCertificateStore(cachedMetadata.CertificatesPEM)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build IdP certificate store: %w", err)
	}

	// Construct gosaml2.SAMLServiceProvider
	samlServiceProvider := &saml2.SAMLServiceProvider{
		// IdP settings (from cached metadata)
		IdentityProviderSSOURL: cachedMetadata.SSOURL,
		IdentityProviderIssuer: cachedMetadata.Issuer,
		IDPCertificateStore:    idpCertStore,

		// SP settings (from tenant config)
		AssertionConsumerServiceURL: fmt.Sprintf("%s/saml/callback", tenantSettings.Endpoint),
		ServiceProviderIssuer:       tenantSettings.Endpoint,
		ServiceProviderSLOURL:       fmt.Sprintf("%s/saml/logout", tenantSettings.Endpoint),
		SPKeyStore:                  spCertStore,

		// Options (from tenant config)
		SignAuthnRequests:       tenantSettings.Options.SignAuthnRequests,
		ForceAuthn:              tenantSettings.Options.ForceLogin,
		IsPassive:               false,
		AudienceURI:             tenantSettings.AudienceUri,
		ValidateEncryptionCert:  tenantSettings.Options.ValidateEncryptionCertificate,
		SkipSignatureValidation: tenantSettings.Options.SkipSignatureValidation,
		AllowMissingAttributes:  tenantSettings.Options.AllowMissingAttributes,
	}

	// Parse attribute map from provider
	var attributeMap samlConfig.AttributeMap
	if provider.AttributeMap != "" {
		err = json.Unmarshal([]byte(provider.AttributeMap), &attributeMap)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal attribute map: %w", err)
		}
	}

	// Apply defaults for empty attribute mappings
	b.useDefaultAttributesIfEmpty(&attributeMap, provider.Name)

	providerConfig := &ProviderConfig{
		AttributeMap:          attributeMap,
		SkipEmailVerification: provider.SkipEmailVerification,
		Domain:                provider.Domain,
	}

	return samlServiceProvider, providerConfig, nil
}

// loadServiceProviderCertificate loads the SP certificate for signing, auto-creating if missing
func (b *ProviderManager) loadServiceProviderCertificate(tenantID uuid.UUID, serviceName string) (dsig.X509KeyStore, error) {
	cert, err := b.persister.GetSamlCertificatePersister().GetFirst(tenantID)
	if err != nil {
		return nil, err
	}

	if cert == nil {
		return nil, fmt.Errorf("no SAML certificate found for tenant %s", tenantID)
	}

	privateKey, err := cert.DecryptCertKey()
	if err != nil {
		return nil, err
	}

	keys, err := tls.X509KeyPair([]byte(cert.CertData), privateKey)
	if err != nil {
		return nil, fmt.Errorf("unable to create key pair: %w", err)
	}

	keys.Leaf, err = x509.ParseCertificate(keys.Certificate[0])
	if err != nil {
		return nil, fmt.Errorf("unable to parse certificate: %w", err)
	}

	return dsig.TLSCertKeyStore(keys), nil
}

// buildIdPCertificateStore creates a certificate store from PEM-encoded certificates
func (b *ProviderManager) buildIdPCertificateStore(certificatesPEM []string) (dsig.X509CertificateStore, error) {
	certStore := dsig.MemoryX509CertificateStore{
		Roots: []*x509.Certificate{},
	}

	for i, pemCert := range certificatesPEM {
		block, _ := pem.Decode([]byte(pemCert))
		if block == nil {
			return nil, fmt.Errorf("failed to decode PEM certificate at index %d", i)
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse certificate at index %d: %w", i, err)
		}

		certStore.Roots = append(certStore.Roots, cert)
	}

	return &certStore, nil
}

// useDefaultAttributesIfEmpty sets default SAML attribute mappings
func (b *ProviderManager) useDefaultAttributesIfEmpty(attributeMap *samlConfig.AttributeMap, providerName string) {
	// Generic defaults for other providers
	if attributeMap.Name == "" {
		attributeMap.Name = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name"
	}
	if attributeMap.GivenName == "" {
		attributeMap.GivenName = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname"
	}
	if attributeMap.FamilyName == "" {
		attributeMap.FamilyName = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname"
	}
	if attributeMap.Email == "" {
		attributeMap.Email = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"
	}

	// Detect Auth0 provider and use Auth0-specific defaults
	if strings.Contains(strings.ToLower(providerName), "auth0") {
		b.useAuth0Defaults(attributeMap)
		return
	}
}

// useAuth0Defaults sets Auth0-specific attribute mappings
func (b *ProviderManager) useAuth0Defaults(attributeMap *samlConfig.AttributeMap) {
	if attributeMap.Name == "" {
		attributeMap.Name = "http://schemas.auth0.com/name"
	}
	//if attributeMap.Email == "" {
	//	attributeMap.Email = "http://schemas.auth0.com/email"
	//}
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
