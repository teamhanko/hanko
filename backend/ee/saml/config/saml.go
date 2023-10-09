package config

import (
	"errors"
	"fmt"
	"github.com/gobwas/glob"
	"net/url"
	"strings"
)

type Saml struct {
	Enabled               bool                 `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=false"`
	Endpoint              string               `yaml:"endpoint_url" json:"endpoint_url,omitempty" koanf:"endpoint_url"`
	AudienceUri           string               `yaml:"audience_uri" json:"audience_uri,omitempty" koanf:"audience_uri"`
	DefaultRedirectUrl    string               `yaml:"default_redirect_url" json:"default_redirect_url,omitempty" koanf:"default_redirect_url"`
	AllowedRedirectURLS   []string             `yaml:"allowed_redirect_urls" json:"allowed_redirect_urls,omitempty" koanf:"allowed_redirect_urls" split_words:"true"`
	AllowedRedirectURLMap map[string]glob.Glob `jsonschema:"-"`

	Options Options `yaml:"options" json:"options,omitempty" koanf:"options"`

	IdentityProviders []IdentityProvider `yaml:"identity_providers" json:"identity_providers,omitempty" koanf:"identity_providers"`
}

type Options struct {
	SignAuthnRequests bool `yaml:"sign_authn_requests" json:"sign_authn_requests,omitempty" koanf:"sign_authn_requests" jsonschema:"default=true"`
	// Forces the IDP to show login window every time
	ForceLogin                    bool `yaml:"force_login" json:"force_login,omitempty" koanf:"force_login" jsonschema:"default=false"`
	ValidateEncryptionCertificate bool `yaml:"validate_encryption_cert" json:"validate_encryption_cert,omitempty" koanf:"validate_encryption_cert" jsonschema:"default=true"`
	SkipSignatureValidation       bool `yaml:"skip_signature_validation" json:"skip_signature_validation,omitempty" koanf:"skip_signature_validation" jsonschema:"default=false"`
	AllowMissingAttributes        bool `yaml:"allow_missing_attributes" json:"allow_missing_attributes,omitempty" koanf:"allow_missing_attributes" jsonschema:"default=false"`
}

type IdentityProvider struct {
	Enabled               bool         `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=false"`
	Name                  string       `yaml:"name" json:"name,omitempty" koanf:"name"`
	Domain                string       `yaml:"domain" json:"domain,omitempty" koanf:"domain"`
	MetadataUrl           string       `yaml:"metadata_url" json:"metadata_url,omitempty" koanf:"metadata_url"`
	SkipEmailVerification bool         `yaml:"skip_email_verification" json:"skip_email_verification,omitempty" koanf:"skip_email_verification"`
	AttributeMap          AttributeMap `yaml:"attribute_map" json:"attribute_map,omitempty" koanf:"attribute_map"`
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
