package config

import (
	"errors"
	"fmt"
	"github.com/gobwas/glob"
	"net/url"
	"strings"
)

type Saml struct {
	// `enabled` determines whether the SAML API endpoints are available.
	Enabled bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=false"`
	// `endpoint` is URL at which the SAML endpoints like metadata, callback, etc. are available
	// (e.g. `{YOUR_BACKEND_INSTANCE}/api`).
	//
	// Will be provided as metadata for IdP.
	Endpoint string `yaml:"endpoint_url" json:"endpoint_url,omitempty" koanf:"endpoint_url"`
	// `audience_uri` determines the intended recipient or audience for the SAML Assertion.
	AudienceUri string `yaml:"audience_uri" json:"audience_uri,omitempty" koanf:"audience_uri"`
	// `default_redirect_url` is the URL to redirect to in case of errors or when no `allowed_redirect_url` is provided.
	DefaultRedirectUrl string `yaml:"default_redirect_url" json:"default_redirect_url,omitempty" koanf:"default_redirect_url"`
	// `allowed_redirect_urls` is a list of URLs the backend is allowed to redirect to after third party sign-in was
	// successful.
	//
	// Supports wildcard matching through globbing. e.g. `https://*.example.com` will allow `https://foo.example.com`
	// and `https://bar.example.com` to be accepted.
	//
	// Globbing is also supported for paths, e.g. `https://foo.example.com/*` will match `https://foo.example.com/page1`
	// and `https://foo.example.com/page2`.
	//
	// A double asterisk (`**`) acts as a "super"-wildcard/match-all.
	//
	// See [here](https://pkg.go.dev/github.com/gobwas/glob#Compile) for more on globbinh.
	AllowedRedirectURLS   []string             `yaml:"allowed_redirect_urls" json:"allowed_redirect_urls,omitempty" koanf:"allowed_redirect_urls" split_words:"true"`
	AllowedRedirectURLMap map[string]glob.Glob `jsonschema:"-"`

	// `options` allows setting optional features for service provider operations.
	Options Options `yaml:"options" json:"options,omitempty" koanf:"options" jsonschema:"title=options"`

	// `identity_providers` is a list of SAML identity providers.
	IdentityProviders []IdentityProvider `yaml:"identity_providers" json:"identity_providers,omitempty" koanf:"identity_providers"`
}

func (s Saml) GetProviderByDomain(domain string) *IdentityProvider {
	for _, ip := range s.IdentityProviders {
		if ip.Domain == domain {
			return &ip
		}
	}

	return nil
}

type Options struct {
	// `sign_authn_requests` determines whether initial requests should be signed.
	SignAuthnRequests bool `yaml:"sign_authn_requests" json:"sign_authn_requests,omitempty" koanf:"sign_authn_requests" jsonschema:"default=true"`
	// `force_login` forces the IdP to always show a login (even if there is an active session with the IdP).
	ForceLogin bool `yaml:"force_login" json:"force_login,omitempty" koanf:"force_login" jsonschema:"default=false"`
	// `validate_encryption_cert` determines whether the certificate used for the encryption of the IdP responses should
	// be checked for validity.
	ValidateEncryptionCertificate bool `yaml:"validate_encryption_cert" json:"validate_encryption_cert,omitempty" koanf:"validate_encryption_cert" jsonschema:"default=true"`
	// `skip_signature_validation` determines whether the validity check of an IdP response's signature
	// should be skipped.
	SkipSignatureValidation bool `yaml:"skip_signature_validation" json:"skip_signature_validation,omitempty" koanf:"skip_signature_validation" jsonschema:"default=false"`
	// `allow_missing_attributes` determines whether missing attributes are allowed (e.g. the IdP specifies a phone
	// attribute in the metadata but does not send it with a SAML Assertion Response).
	AllowMissingAttributes bool `yaml:"allow_missing_attributes" json:"allow_missing_attributes,omitempty" koanf:"allow_missing_attributes" jsonschema:"default=false"`
}

type IdentityProvider struct {
	// `enabled` activates or deactivates the identity provider.
	Enabled bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=false"`
	// `name` is the name given for the identity provider.
	Name string `yaml:"name" json:"name,omitempty" koanf:"name"`
	// At login the domain will be extracted from the users email address and then used to identify the idp to use.
	// This tag defines for which domain the idp is used.
	Domain string `yaml:"domain" json:"domain,omitempty" koanf:"domain"`
	// `metadata_url` is the URL the API can retrieve IdP metadata from.
	MetadataUrl string `yaml:"metadata_url" json:"metadata_url,omitempty" koanf:"metadata_url"`
	// `skip_email_verification` determines whether the check if the `email_verified` attribute in the IdP response
	// will be skipped.
	SkipEmailVerification bool `yaml:"skip_email_verification" json:"skip_email_verification,omitempty" koanf:"skip_email_verification"`
	// `attribute_map` is a map of attributes used to map attributes in IdP response to custom attributes at
	// Hanko.
	AttributeMap AttributeMap `yaml:"attribute_map" json:"attribute_map,omitempty" koanf:"attribute_map" jsonschema:"title=attribute_map"`
}

type AttributeMap struct {
	Name              string `yaml:"name" json:"name,omitempty" koanf:"name" jsonschema:"default=http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name"`
	FamilyName        string `yaml:"family_name" json:"family_name,omitempty" koanf:"family_name" jsonschema:"default=http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname"`
	GivenName         string `yaml:"given_name" json:"given_name,omitempty" koanf:"given_name" jsonschema:"default=http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname"`
	MiddleName        string `yaml:"middle_name" json:"middle_name,omitempty" koanf:"middle_name"`
	NickName          string `yaml:"nickname" json:"nickname,omitempty" koanf:"nickname"`
	PreferredUsername string `yaml:"preferred_username" json:"preferred_username,omitempty" koanf:"preferred_username"`
	Profile           string `yaml:"profile" json:"profile,omitempty" koanf:"profile"`
	Picture           string `yaml:"picture" json:"picture,omitempty" koanf:"picture"`
	Website           string `yaml:"website" json:"website,omitempty" koanf:"website"`
	Gender            string `yaml:"gender" json:"gender,omitempty" koanf:"gender"`
	Birthdate         string `yaml:"birthdate" json:"birthdate,omitempty" koanf:"birthdate"`
	ZoneInfo          string `yaml:"zone_info" json:"zone_info,omitempty" koanf:"zone_info"`
	Locale            string `yaml:"locale" json:"locale,omitempty" koanf:"locale"`
	UpdatedAt         string `yaml:"updated_at" json:"updated_at,omitempty" koanf:"updated_at"`
	Email             string `yaml:"email" json:"email,omitempty" koanf:"email" jsonschema:"default=http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"`
	EmailVerified     string `yaml:"email_verified" json:"email_verified,omitempty" koanf:"email_verified"`
	Phone             string `yaml:"phone" json:"phone,omitempty" koanf:"phone"`
	PhoneVerified     string `yaml:"phone_verified" json:"phone_verified,omitempty" koanf:"phone_verified"`
}

func (s *Saml) PostProcess() error {
	s.Endpoint = strings.TrimSuffix(s.Endpoint, "/")

	s.AllowedRedirectURLMap = make(map[string]glob.Glob)
	urls := append(s.AllowedRedirectURLS, s.DefaultRedirectUrl)
	for _, redirectUrl := range urls {
		globbedUrl, err := glob.Compile(redirectUrl, '.', '/')
		if err != nil {
			return fmt.Errorf("failed compile allowed redirect url glob: %w", err)
		}
		s.AllowedRedirectURLMap[redirectUrl] = globbedUrl
	}

	return nil
}

func (s *Saml) Validate() error {
	if s.Enabled {
		validationErrors := s.ValidateEmpty()
		if validationErrors != nil {
			return validationErrors
		}

		validationErrors = s.ValidateUrls()
		if validationErrors != nil {
			return validationErrors
		}

		if len(s.IdentityProviders) == 0 {
			return errors.New("at least one SAML provider is needed")
		}

		for _, provider := range s.IdentityProviders {
			validationErrors = provider.Validate()
			if validationErrors != nil {
				return validationErrors
			}
		}
	}

	return nil
}

func (s *Saml) ValidateEmpty() error {
	if strings.TrimSpace(s.Endpoint) == "" {
		return errors.New("endpoint_url must be set")
	}

	if strings.TrimSpace(s.AudienceUri) == "" {
		return errors.New("audience_uri must be set")
	}

	return nil
}

const (
	invalidUrlFormat = "'%s' is not a valid url"
)

func (s *Saml) ValidateUrls() error {
	_, err := url.Parse(s.Endpoint)
	if err != nil {
		return fmt.Errorf(invalidUrlFormat, s.Endpoint)
	}

	_, err = url.Parse(s.DefaultRedirectUrl)
	if err != nil {
		return fmt.Errorf(invalidUrlFormat, s.DefaultRedirectUrl)
	}

	for _, redirectUrl := range s.AllowedRedirectURLS {
		_, err = url.Parse(s.DefaultRedirectUrl)
		if err != nil {
			return fmt.Errorf(invalidUrlFormat, redirectUrl)
		}
	}

	return nil
}

func (idp *IdentityProvider) Validate() error {
	if strings.TrimSpace(idp.Domain) == "" {
		return errors.New("domain must be set")
	}

	if strings.TrimSpace(idp.Name) == "" {
		return errors.New("name must be set")
	}

	_, err := url.Parse(fmt.Sprintf("http://%s", idp.Domain))
	if err != nil {
		return fmt.Errorf(invalidUrlFormat, idp.Domain)
	}

	if strings.TrimSpace(idp.MetadataUrl) == "" {
		return errors.New("identity provider metadata url must be set")
	}

	_, err = url.Parse(idp.MetadataUrl)
	if err != nil {
		return fmt.Errorf(invalidUrlFormat, idp.MetadataUrl)
	}

	return nil
}
