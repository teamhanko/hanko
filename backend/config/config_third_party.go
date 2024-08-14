package config

import (
	"errors"
	"fmt"
	"github.com/fatih/structs"
	"github.com/gobwas/glob"
	"github.com/invopop/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"strings"
)

type ThirdParty struct {
	// `providers` contains the configurations for the available OAuth/OIDC identity providers.
	Providers ThirdPartyProviders `yaml:"providers" json:"providers,omitempty" koanf:"providers" jsonschema:"title=providers,uniqueItems=true"`
	// `redirect_url` is the URL the third party provider redirects to with an authorization code. Must consist of the base URL
	// of your running Hanko backend instance and the `callback` endpoint of the API,
	// i.e. `{YOUR_BACKEND_INSTANCE}/thirdparty/callback.`
	//
	// Required if any of the [`providers`](#providers) are `enabled`.
	RedirectURL string `yaml:"redirect_url" json:"redirect_url,omitempty" koanf:"redirect_url" split_words:"true" jsonschema:"example=https://yourinstance.com/thirdparty/callback"`
	// `error_redirect_url` is the URL the backend redirects to if an error occurs during third party sign-in.
	// Errors are provided as 'error' and 'error_description' query params in the redirect location URL.
	//
	// When using the Hanko web components it should be the URL of the page that embeds the web component such that
	// errors can be processed properly by the web component.
	//
	// You do not have to add this URL to the 'allowed_redirect_urls', it is automatically included when validating
	// redirect URLs.
	//
	// Required if any of the [`providers`](#providers) are `enabled`. Must not have trailing slash.
	ErrorRedirectURL string `yaml:"error_redirect_url" json:"error_redirect_url,omitempty" koanf:"error_redirect_url" split_words:"true"`
	// `default_redirect_url` is the URL the backend redirects to after it successfully verified
	// the response from any third party provider.
	//
	// Must not have trailing slash.
	DefaultRedirectURL string `yaml:"default_redirect_url" json:"default_redirect_url,omitempty" koanf:"default_redirect_url" split_words:"true"`
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
	// See [here](https://pkg.go.dev/github.com/gobwas/glob#Compile) for more on globbing.
	//
	// Must not be empty if any of the [`providers`](#providers) are `enabled`. URLs in the list must not have a trailing slash.
	AllowedRedirectURLS   []string             `yaml:"allowed_redirect_urls" json:"allowed_redirect_urls,omitempty" koanf:"allowed_redirect_urls" split_words:"true"`
	AllowedRedirectURLMap map[string]glob.Glob `jsonschema:"-"`
}

func (t *ThirdParty) Validate() error {
	if t.Providers.HasEnabled() {
		if t.RedirectURL == "" {
			return errors.New("redirect_url must be set")
		}

		if t.ErrorRedirectURL == "" {
			return errors.New("error_redirect_url must be set")
		}

		if len(t.AllowedRedirectURLS) <= 0 {
			return errors.New("at least one allowed redirect url must be set")
		}

		urls := append(t.AllowedRedirectURLS, t.ErrorRedirectURL)
		if t.DefaultRedirectURL != "" {
			urls = append(urls, t.DefaultRedirectURL)
		}
		for _, u := range urls {
			if strings.HasSuffix(u, "/") {
				return fmt.Errorf("redirect url %s must not have trailing slash", u)
			}
		}
	}

	err := t.Providers.Validate()
	if err != nil {
		return fmt.Errorf("failed to validate third party providers: %w", err)
	}

	return nil
}

func (t *ThirdParty) PostProcess() error {
	t.AllowedRedirectURLMap = make(map[string]glob.Glob)
	urls := append(t.AllowedRedirectURLS, t.ErrorRedirectURL)
	for _, redirectUrl := range urls {
		g, err := glob.Compile(redirectUrl, '.', '/')
		if err != nil {
			return fmt.Errorf("failed compile allowed redirect url glob: %w", err)
		}
		t.AllowedRedirectURLMap[redirectUrl] = g
	}

	return nil
}

type ThirdPartyProviders struct {
	// `apple` contains the provider configuration for Apple.
	Apple ThirdPartyProvider `yaml:"apple" json:"apple,omitempty" koanf:"apple"`
	// `discord` contains the provider configuration for Discord.
	Discord ThirdPartyProvider `yaml:"discord" json:"discord,omitempty" koanf:"discord"`
	// `github` contains the provider configuration for GitHub.
	GitHub ThirdPartyProvider `yaml:"github" json:"github,omitempty" koanf:"github"`
	// `google` contains the provider configuration for Google.
	Google ThirdPartyProvider `yaml:"google" json:"google,omitempty" koanf:"google"`
	// `linkedin` contains the provider configuration for LinkedIn.
	LinkedIn ThirdPartyProvider `yaml:"linkedin" json:"linkedin,omitempty" koanf:"linkedin"`
	// `microsoft` contains the provider configuration for Microsoft.
	Microsoft ThirdPartyProvider `yaml:"microsoft" json:"microsoft,omitempty" koanf:"microsoft"`
}

func (p *ThirdPartyProviders) Validate() error {
	s := structs.New(p)
	for _, field := range s.Fields() {
		provider := field.Value().(ThirdPartyProvider)
		err := provider.Validate()
		if err != nil {
			return fmt.Errorf("%s: %w", strings.ToLower(field.Name()), err)
		}
	}
	return nil
}

func (p *ThirdPartyProviders) HasEnabled() bool {
	s := structs.New(p)
	for _, field := range s.Fields() {
		provider := field.Value().(ThirdPartyProvider)
		if provider.Enabled {
			return true
		}
	}

	return false
}

func (p *ThirdPartyProviders) GetEnabled() []ThirdPartyProvider {
	s := structs.New(p)
	var enabledProviders []ThirdPartyProvider
	for _, field := range s.Fields() {
		provider := field.Value().(ThirdPartyProvider)
		if provider.Enabled {
			enabledProviders = append(enabledProviders, provider)
		}
	}

	return enabledProviders
}

func (p *ThirdPartyProviders) Get(provider string) *ThirdPartyProvider {
	s := structs.New(p)
	for _, field := range s.Fields() {
		if strings.ToLower(field.Name()) == strings.ToLower(provider) {
			p := field.Value().(ThirdPartyProvider)
			return &p
		}
	}

	return nil
}

type ThirdPartyProvider struct {
	// `allow_linking` indicates whether existing accounts can be automatically linked with this provider.
	//
	// Linking is based on matching one of the email addresses of an existing user account with the (primary)
	// email address of the third party provider account.
	AllowLinking bool `yaml:"allow_linking" json:"allow_linking,omitempty" koanf:"allow_linking" split_words:"true"`
	// `client_id` is the ID of the OAuth/OIDC client. Must be obtained from the provider.
	//
	// Required if the provider is `enabled`.
	ClientID    string `yaml:"client_id" json:"client_id,omitempty" koanf:"client_id" split_words:"true"`
	DisplayName string `jsonschema:"-" yaml:"-" json:"-" koanf:"-"`
	// `enabled` determines whether this provider is enabled.
	Enabled bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=false"`
	// `secret` is the client secret for the OAuth/OIDC client. Must be obtained from the provider.
	//
	// Required if the provider is `enabled`.
	Secret string `yaml:"secret" json:"secret,omitempty" koanf:"secret"`
}

func (ThirdPartyProvider) JSONSchemaExtend(schema *jsonschema.Schema) {
	schema.Title = "provider"

	enabledTrue := &jsonschema.Schema{Properties: orderedmap.New[string, *jsonschema.Schema]()}
	enabledTrue.Properties.Set("enabled", &jsonschema.Schema{Const: true})

	schema.If = enabledTrue
	schema.Then = &jsonschema.Schema{
		Required: []string{"client_id", "secret"},
	}
	schema.Else = &jsonschema.Schema{
		Required: []string{"enabled"},
	}
}

func (p *ThirdPartyProvider) Validate() error {
	if p.Enabled {
		if p.ClientID == "" {
			return errors.New("missing client ID")
		}
		if p.Secret == "" {
			return errors.New("missing client secret")
		}
	}
	return nil
}
