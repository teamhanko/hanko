package provider

import (
	"encoding/xml"
	"fmt"
	"github.com/fatih/structs"
	saml2 "github.com/russellhaering/gosaml2"
	"github.com/teamhanko/hanko/backend/config"
	samlConfig "github.com/teamhanko/hanko/backend/ee/saml/config"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/thirdparty"
	"strings"
	"time"
)

type BaseSamlProvider struct {
	Config  samlConfig.IdentityProvider
	Service saml2.SAMLServiceProvider
}

func NewBaseSamlProvider(cfg *config.Config, idpConfig samlConfig.IdentityProvider, persister persistence.SamlCertificatePersister) (ServiceProvider, error) {
	serviceProviderCertStore, err := loadCertificate(cfg, persister)
	if err != nil {
		return nil, err
	}

	idpMetadata, err := fetchIdpMetadata(idpConfig)
	if err != nil {
		return nil, err
	}

	provider := &BaseSamlProvider{
		Config: idpConfig,
		Service: saml2.SAMLServiceProvider{
			IdentityProviderSSOURL: idpMetadata.SingleSignOnUrl,
			IdentityProviderIssuer: idpMetadata.Issuer,
			IDPCertificateStore:    &idpMetadata.certs,

			AssertionConsumerServiceURL: fmt.Sprintf("%s/saml/callback", cfg.Saml.Endpoint),
			ServiceProviderIssuer:       fmt.Sprintf("%s/saml/metadata", cfg.Saml.Endpoint),
			ServiceProviderSLOURL:       fmt.Sprintf("%s/saml/logout", cfg.Saml.Endpoint),
			SPKeyStore:                  serviceProviderCertStore,

			SignAuthnRequests:       cfg.Saml.Options.SignAuthnRequests,
			ForceAuthn:              cfg.Saml.Options.ForceLogin,
			IsPassive:               false,
			AudienceURI:             cfg.Saml.AudienceUri,
			ValidateEncryptionCert:  cfg.Saml.Options.ValidateEncryptionCertificate,
			SkipSignatureValidation: cfg.Saml.Options.SkipSignatureValidation,
			AllowMissingAttributes:  cfg.Saml.Options.AllowMissingAttributes,
		},
	}
	provider.UseDefaultAttributesIfEmpty()

	return provider, nil
}

func (sp *BaseSamlProvider) ProvideMetadataAsXml() ([]byte, error) {
	metadata, err := sp.Service.Metadata()
	if err != nil {
		return nil, err
	}

	// Workaround as the lib is currently marshalling nanoseconds which cannot be used
	metadata.ValidUntil = time.Now().Add(time.Hour * 24 * 7).Round(time.Millisecond)

	return xml.MarshalIndent(metadata, "", "  ")
}

func (sp *BaseSamlProvider) GetUserData(assertionInfo *saml2.AssertionInfo) *thirdparty.UserData {
	firstAssertion := assertionInfo.Assertions[0]
	assertionValues := assertionInfo.Values
	attributeMap := &sp.Config.AttributeMap

	emailAddress := assertionValues.Get(attributeMap.Email)

	email := thirdparty.Email{
		Email:    emailAddress,
		Verified: false,
		Primary:  true,
	}

	expiresIn, _ := time.Parse(time.RFC3339, firstAssertion.Conditions.NotOnOrAfter)

	userData := &thirdparty.UserData{}
	userData.Emails = append(userData.Emails, email)

	userData.Metadata = &thirdparty.Claims{
		Issuer:            firstAssertion.Issuer.Value,
		Subject:           firstAssertion.Subject.NameID.Value,
		Aud:               sp.Service.AudienceURI,
		Iat:               float64(assertionInfo.AuthnInstant.Unix()),
		Exp:               float64(expiresIn.Unix()),
		Name:              assertionValues.Get(attributeMap.Name),
		FamilyName:        assertionValues.Get(attributeMap.FamilyName),
		GivenName:         assertionValues.Get(attributeMap.GivenName),
		MiddleName:        assertionValues.Get(attributeMap.MiddleName),
		NickName:          assertionValues.Get(attributeMap.NickName),
		PreferredUsername: assertionValues.Get(attributeMap.PreferredUsername),
		Profile:           assertionValues.Get(attributeMap.Profile),
		Picture:           assertionValues.Get(attributeMap.Picture),
		Website:           assertionValues.Get(attributeMap.Website),
		Gender:            assertionValues.Get(attributeMap.Gender),
		Birthdate:         assertionValues.Get(attributeMap.Birthdate),
		ZoneInfo:          assertionValues.Get(attributeMap.ZoneInfo),
		Locale:            assertionValues.Get(attributeMap.Locale),
		UpdatedAt:         assertionValues.Get(attributeMap.UpdatedAt),
		Email:             emailAddress,
		EmailVerified:     assertionValues.Get(attributeMap.EmailVerified) != "" || sp.Config.SkipEmailVerification,
		Phone:             assertionValues.Get(attributeMap.Phone),
		PhoneVerified:     assertionValues.Get(attributeMap.PhoneVerified) != "",
		CustomClaims:      sp.mapCustomClaims(assertionInfo.Values, attributeMap),
	}

	return userData
}

func (sp *BaseSamlProvider) mapCustomClaims(values saml2.Values, attributeMap *samlConfig.AttributeMap) map[string]interface{} {
	customAttributes := make(map[string]interface{})

	s := structs.New(attributeMap)
	for name := range values {
		hasField := false
		for _, field := range s.Fields() {
			if field.Value().(string) == name {
				hasField = true
			}
		}

		if !hasField {
			customAttributes[name] = values.Get(name)
		}
	}

	return customAttributes
}

func (sp *BaseSamlProvider) UseDefaultAttributesIfEmpty() {
	attributeMap := &sp.Config.AttributeMap
	if strings.TrimSpace(attributeMap.Name) == "" {
		attributeMap.Name = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name"
	}

	if strings.TrimSpace(attributeMap.GivenName) == "" {
		attributeMap.GivenName = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname"
	}

	if strings.TrimSpace(attributeMap.FamilyName) == "" {
		attributeMap.FamilyName = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname"
	}

	if strings.TrimSpace(attributeMap.Email) == "" {
		attributeMap.Email = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"
	}
}

func (sp *BaseSamlProvider) GetDomain() string {
	return sp.Config.Domain
}

func (sp *BaseSamlProvider) GetService() *saml2.SAMLServiceProvider {
	return &sp.Service
}

func (sp *BaseSamlProvider) GetConfig() samlConfig.IdentityProvider {
	return sp.Config
}
