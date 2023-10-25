package config

import (
	"github.com/gobwas/glob"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSamlConfig_PostProcess(t *testing.T) {
	redirectUrls := []string{"http://allowed-redirect/lorem"}

	cfg := &Saml{
		Endpoint:            "http://lorem.ipsum/",
		DefaultRedirectUrl:  "http://localhost:8000",
		AllowedRedirectURLS: redirectUrls,
	}
	globbedUrl, err := glob.Compile(cfg.DefaultRedirectUrl, '.', '/')
	assert.NoError(t, err)

	err = cfg.PostProcess()
	assert.NoError(t, err)
	assert.NotNil(t, cfg.AllowedRedirectURLMap)
	assert.Len(t, cfg.AllowedRedirectURLMap, 2)
	assert.Equal(t, globbedUrl, cfg.AllowedRedirectURLMap[cfg.DefaultRedirectUrl])
}

func TestSamlConfig_PostProcessWithBrokenGlob(t *testing.T) {
	redirectUrls := []string{"http://allowed-redirect/lorem"}

	cfg := &Saml{
		Endpoint:            "http://lorem.ipsum/",
		DefaultRedirectUrl:  "[",
		AllowedRedirectURLS: redirectUrls,
	}

	err := cfg.PostProcess()
	assert.NotNil(t, err)
	assert.ErrorContains(t, err, "failed compile allowed redirect url glob")
}

func TestSamlConfig_PostProcessEndpointTrimme(t *testing.T) {
	redirectUrls := []string{"http://allowed-redirect/lorem"}

	cfg := &Saml{
		Endpoint:            "http://lorem.ipsum/",
		DefaultRedirectUrl:  "http://localhost:8000",
		AllowedRedirectURLS: redirectUrls,
	}

	err := cfg.PostProcess()
	assert.NoError(t, err)
	assert.Equal(t, "http://lorem.ipsum", cfg.Endpoint)
}

func TestSamlConfig_ValidateWithDisabledSaml(t *testing.T) {
	cfg := &Saml{
		Enabled: false,
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestSamlConfig_ValidateWithEnabledSaml(t *testing.T) {
	cfg := &Saml{
		Enabled:               true,
		Endpoint:              "http://lorem.ipsum/",
		AudienceUri:           "urn:Hanko",
		DefaultRedirectUrl:    "http://lorem.ipsum/saml",
		AllowedRedirectURLS:   []string{},
		AllowedRedirectURLMap: nil,
		IdentityProviders: []IdentityProvider{{
			Enabled:     true,
			Name:        "Test Provider",
			Domain:      "lorem.ipsum",
			MetadataUrl: "http://provider.lorem.ipsum/metadata",
		}},
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestSamlConfig_ValidateEmpty(t *testing.T) {
	cfg := &Saml{
		Endpoint:    "lorem",
		AudienceUri: "ipsum",
	}

	err := cfg.ValidateEmpty()
	assert.NoError(t, err)
}

func TestSamlConfig_ValidateEmptyErrorWithEmptyEndpoint(t *testing.T) {
	cfg := &Saml{
		Endpoint: "",
	}

	err := cfg.ValidateEmpty()
	assert.Equal(t, "endpoint_url must be set", err.Error())
}

func TestSamlConfig_ValidateEmptyErrorWithSpaceEndpoint(t *testing.T) {
	cfg := &Saml{
		Endpoint: "  ",
	}

	err := cfg.ValidateEmpty()
	assert.Equal(t, "endpoint_url must be set", err.Error())
}

func TestSamlConfig_ValidateEmptyErrorWithEmptyAudienceUri(t *testing.T) {
	cfg := &Saml{
		Endpoint:    "http://lorem.ipsum",
		AudienceUri: "",
	}

	err := cfg.ValidateEmpty()
	assert.Equal(t, "audience_uri must be set", err.Error())
}

func TestSamlConfig_ValidateEmptyErrorWithSpaceAudienceUri(t *testing.T) {
	cfg := &Saml{
		Endpoint:    "http://lorem.ipsum",
		AudienceUri: "  ",
	}

	err := cfg.ValidateEmpty()
	assert.Equal(t, "audience_uri must be set", err.Error())
}

func TestSamlConfig_ValidateUrls(t *testing.T) {
	cfg := &Saml{
		Endpoint:            "http://lorem.ipsum",
		DefaultRedirectUrl:  "http://lorem.ipsum/redirect",
		AllowedRedirectURLS: []string{"http://lorem.ipsum/allowed"},
	}

	err := cfg.ValidateUrls()
	assert.NoError(t, err)
}

func TestSamlConfig_ValidateUrlsWithWrongEndpointUrl(t *testing.T) {
	cfg := &Saml{
		Endpoint: "http://lorem:8000.de/ipsum",
	}

	err := cfg.ValidateUrls()
	assert.Errorf(t, err, invalidUrlFormat, cfg.Endpoint)
}

func TestSamlConfig_ValidateUrlsWithWrongDefaultRedirecttUrl(t *testing.T) {
	cfg := &Saml{
		Endpoint:           "http://lorem.ipsum",
		DefaultRedirectUrl: "http://lorem:8000.de/ipsum",
	}

	err := cfg.ValidateUrls()
	assert.Errorf(t, err, invalidUrlFormat, cfg.DefaultRedirectUrl)
}

func TestSamlConfig_ValidateUrlsWithWrongAllowedRedirectUrl(t *testing.T) {
	cfg := &Saml{
		Endpoint:            "http://lorem.ipsum",
		DefaultRedirectUrl:  "http://lorem:8000.de/ipsum",
		AllowedRedirectURLS: []string{"s:/"},
	}

	err := cfg.ValidateUrls()
	assert.Errorf(t, err, invalidUrlFormat, cfg.AllowedRedirectURLS[0])
}

func TestSamlConfig_ValidateProvider(t *testing.T) {
	cfg := &IdentityProvider{
		Enabled:     true,
		Name:        "Lorem",
		Domain:      "lorem.ipsum",
		MetadataUrl: "http://provider.lorem.ipsum/metadata",
	}

	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestSamlConfig_ValidateProviderErrorWithEmnptyDomain(t *testing.T) {
	cfg := &IdentityProvider{
		Domain: "",
	}

	err := cfg.Validate()
	assert.ErrorContains(t, err, "domain must be set")
}

func TestSamlConfig_ValidateProviderErrorWithSpaceDomain(t *testing.T) {
	cfg := &IdentityProvider{
		Domain: "  ",
	}

	err := cfg.Validate()
	assert.ErrorContains(t, err, "domain must be set")
}

func TestSamlConfig_ValidateProviderErrorWithEmptyName(t *testing.T) {
	cfg := &IdentityProvider{
		Domain: "lorem.ipsum",
		Name:   "",
	}

	err := cfg.Validate()
	assert.ErrorContains(t, err, "name must be set")
}

func TestSamlConfig_ValidateProviderErrorWithSpaceName(t *testing.T) {
	cfg := &IdentityProvider{
		Domain: "lorem.ipsum",
		Name:   "  ",
	}

	err := cfg.Validate()
	assert.ErrorContains(t, err, "name must be set")
}

func TestSamlConfig_ValidateProviderErrorWithInvalidDomain(t *testing.T) {
	cfg := &IdentityProvider{
		Domain: "lorem..ipsum",
		Name:   "Test",
	}

	err := cfg.Validate()
	assert.Errorf(t, err, invalidUrlFormat, cfg.Domain)
}

func TestSamlConfig_ValidateProviderErrorWithEmptyMetadataUrl(t *testing.T) {
	cfg := &IdentityProvider{
		Domain:      "lorem.ipsum",
		Name:        "Test",
		MetadataUrl: "",
	}

	err := cfg.Validate()
	assert.ErrorContains(t, err, "identity provider metadata url must be set")
}

func TestSamlConfig_ValidateProviderErrorWithSpaceMetadataUrl(t *testing.T) {
	cfg := &IdentityProvider{
		Domain:      "lorem.ipsum",
		Name:        "Test",
		MetadataUrl: "  ",
	}

	err := cfg.Validate()
	assert.ErrorContains(t, err, "identity provider metadata url must be set")
}

func TestSamlConfig_ValidateProviderErrorWithInvalidMetadataUrl(t *testing.T) {
	cfg := &IdentityProvider{
		Domain:      "lorem.ipsum",
		Name:        "Test",
		MetadataUrl: "http://lorem:8000.de/ipsum",
	}

	err := cfg.Validate()
	assert.Errorf(t, err, invalidUrlFormat, cfg.MetadataUrl)
}
