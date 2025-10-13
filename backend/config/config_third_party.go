package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fatih/structs"
	"github.com/gobwas/glob"
	"github.com/invopop/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type ThirdParty struct {
	// `providers` contains the configurations for the available OAuth/OIDC identity providers.
	Providers ThirdPartyProviders `yaml:"providers" json:"providers,omitempty" koanf:"providers" jsonschema:"title=providers,uniqueItems=true"`
	// `custom_providers contains the configurations for custom OAuth/OIDC identity providers.
	CustomProviders CustomThirdPartyProviders `yaml:"custom_providers" json:"custom_providers,omitempty" koanf:"custom_providers" jsonschema:"title=custom_providers"`
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
	AllowedRedirectURLS   []string             `yaml:"allowed_redirect_urls" json:"allowed_redirect_urls,omitempty" koanf:"allowed_redirect_urls" split_words:"true" jsonschema:"minItems=1"`
	AllowedRedirectURLMap map[string]glob.Glob `jsonschema:"-" yaml:"-" json:"-" koanf:"-"`
}

func (t *ThirdParty) Validate() error {
	hasEnabledProviders := t.Providers.HasEnabled()
	hasEnabledCustomProviders := t.CustomProviders.HasEnabled()

	if hasEnabledProviders || hasEnabledCustomProviders {
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

	if hasEnabledProviders {
		err := t.Providers.Validate()
		if err != nil {
			return fmt.Errorf("failed to validate third party providers: %w", err)
		}
	}

	if hasEnabledCustomProviders {
		err := t.CustomProviders.Validate()
		if err != nil {
			return fmt.Errorf("failed to validate custom third party providers: %w", err)
		}
	}

	return nil
}

func (t ThirdParty) JSONSchemaExtend(schema *jsonschema.Schema) {
	schema.If = &jsonschema.Schema{
		AllOf: []*jsonschema.Schema{
			t.JSONSchemaNoBuiltInProviderEnabled(),
			t.JSONSchemaNoCustomProviderEnabled(),
		},
	}
	schema.Then = &jsonschema.Schema{
		Required: []string{},
	}
	schema.Else = &jsonschema.Schema{
		Required: []string{"redirect_url", "error_redirect_url", "allowed_redirect_urls"},
	}
}

func (t ThirdParty) JSONSchemaNoBuiltInProviderEnabled() *jsonschema.Schema {
	enabledFalseOrNullProperties := orderedmap.New[string, *jsonschema.Schema]()
	enabledFalseOrNullProperties.Set("enabled", &jsonschema.Schema{
		AnyOf: []*jsonschema.Schema{
			{Const: false},
			{Const: "null"},
		},
	})
	enabledFalseOrNullSchema := &jsonschema.Schema{
		Type: "object",
		PatternProperties: map[string]*jsonschema.Schema{
			"^.*": {Ref: "#/$defs/ThirdPartyProvider", Properties: enabledFalseOrNullProperties},
		},
	}

	properties := orderedmap.New[string, *jsonschema.Schema]()
	properties.Set("providers", enabledFalseOrNullSchema)

	return &jsonschema.Schema{Properties: properties}
}

func (t ThirdParty) JSONSchemaNoCustomProviderEnabled() *jsonschema.Schema {
	enabledFalseOrNullProperties := orderedmap.New[string, *jsonschema.Schema]()
	enabledFalseOrNullProperties.Set("enabled", &jsonschema.Schema{
		AnyOf: []*jsonschema.Schema{
			{Const: false},
			{Type: "null"},
		},
	})
	enabledFalseOrNullSchema := &jsonschema.Schema{
		AdditionalProperties: &jsonschema.Schema{
			Ref: "#/$defs/CustomThirdPartyProvider", Properties: enabledFalseOrNullProperties,
		},
	}

	properties := orderedmap.New[string, *jsonschema.Schema]()
	properties.Set("custom_providers", enabledFalseOrNullSchema)

	return &jsonschema.Schema{Properties: properties}
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

	if t.CustomProviders != nil {
		providers := make(map[string]CustomThirdPartyProvider)
		for key, provider := range t.CustomProviders {
			// add prefix per default to ensure built-in and custom providers can be distinguished
			keyLower := strings.ToLower(key)
			provider.ID = "custom_" + keyLower
			providers[keyLower] = provider
		}
		t.CustomProviders = providers
	}

	return nil
}

type CustomThirdPartyProviders map[string]CustomThirdPartyProvider

func (p *CustomThirdPartyProviders) GetEnabled() []CustomThirdPartyProvider {
	var enabledProviders []CustomThirdPartyProvider
	for _, provider := range *p {
		if provider.Enabled {
			enabledProviders = append(enabledProviders, provider)
		}
	}

	return enabledProviders
}

func (p *CustomThirdPartyProviders) HasEnabled() bool {
	for _, provider := range *p {
		if provider.Enabled {
			return true
		}
	}

	return false
}

func (p *CustomThirdPartyProviders) Validate() error {
	for _, v := range p.GetEnabled() {
		err := v.Validate()
		if err != nil {
			return fmt.Errorf(
				"failed to validate third party provider %s: %w",
				strings.TrimPrefix(v.ID, "custom_"),
				err,
			)
		}
	}
	return nil
}

type CustomThirdPartyProvider struct {
	// `allow_linking` indicates whether existing accounts can be automatically linked with this provider.
	//
	// Linking is based on matching one of the email addresses of an existing user account with the (primary)
	// email address of the third party provider account.
	AllowLinking bool `yaml:"allow_linking" json:"allow_linking,omitempty" koanf:"allow_linking" jsonschema:"default=false"`
	// `attribute_mapping` defines a map that associates a set of known standard OIDC conformant end-user claims
	// (the key of a map entry) at the Hanko backend to claims retrieved from a third party provider (the value of the
	// map entry). This is primarily necessary if a non-OIDC provider is configured/used in which case it is probable
	// that user data returned from the userinfo endpoint does not already conform to OIDC standard claims.
	//
	// Example: You configure an OAuth Provider (i.e. non-OIDC) and the provider's configured userinfo endpoint returns
	// an end-user's user ID at the provider not under a `sub` key in its JSON response but rather under a `user_id`
	// key. You would then configure an attribute mapping as follows:
	//
	// ```yaml
	//attribute_mapping:
	//  sub: user_id
	// ```
	//
	// See https://openid.net/specs/openid-connect-core-1_0.html#StandardClaims for a list of known standard claims
	// that provider claims can be mapped into. Any other claims received from a provider are not discarded but are
	// retained internally in a `custom_claims` claim.
	//
	// Mappings are one-to-one mappings, complex mappings (e.g. mapping concatenations of two claims) are not possible.
	AttributeMapping map[string]string `yaml:"attribute_mapping" json:"attribute_mapping,omitempty" koanf:"attribute_mapping"`
	// URL of the provider's authorization endpoint where the end-user is redirected to authenticate and grant consent for
	// an application to access their resources.
	//
	// Required if `use_discovery` is false or omitted.
	AuthorizationEndpoint string `yaml:"authorization_endpoint" json:"authorization_endpoint,omitempty" koanf:"authorization_endpoint"`
	// `ID` is a unique identifier for the provider, derived from the key in the `custom_providers` map, by
	// concatenating the prefix "custom_". This allows distinguishing between built-in and custom providers at runtime.
	ID string `jsonschema:"-" yaml:"-" json:"-" koanf:"-"`
	// `issuer` is the provider's issuer identifier. It should be a URL that uses the "https"
	//	scheme and has no query or fragment components.
	//
	// Required if `use_discovery` is true.
	Issuer string `yaml:"issuer" json:"issuer,omitempty" koanf:"issuer"`
	// `client_id` is the ID of the OAuth/OIDC client. Must be obtained from the provider.
	//
	// Required if the provider is `enabled`.
	ClientID string `yaml:"client_id" json:"client_id,omitempty" koanf:"client_id" split_words:"true"`
	// `display_name` is the name of the provider that is intended to be shown to an end-user.
	//
	// Required if the provider is `enabled`.
	DisplayName string `yaml:"display_name" json:"display_name,omitempty" koanf:"display_name"`
	// `enabled` indicates if the provider is enabled or disabled.
	Enabled bool `yaml:"enabled" json:"enabled,omitempty" koanf:"enabled" jsonschema:"default=false"`
	// `prompt` specifies whether the Authorization Server prompts the End-User for reauthentication and consent.
	// Possible values are:
	// - login
	// - none
	// - consent
	// - select_account
	// Please note that not all providers support all values. Check the corresponding docs of the provider for supported values.
	Prompt string `yaml:"prompt" json:"prompt,omitempty" koanf:"prompt" jsonschema:"default=consent"`
	// `scopes` is a list of scopes requested from the provider that specify the level of access an application has to
	// a user's resources on a server, defining what actions the app can perform on behalf of the user.
	//
	// Required if the provider is `enabled`.
	Scopes []string `yaml:"scopes" json:"scopes,omitempty" koanf:"scopes,omitempty"`
	// `secret` is the client secret for the OAuth/OIDC client. Must be obtained from the provider.
	//
	// Required if the provider is `enabled`.
	Secret string `yaml:"secret" json:"secret,omitempty" koanf:"secret"`
	// URL of the provider's token endpoint URL where an application exchanges an authorization code for an access
	// token, which is used to authenticate API requests on behalf of the end-user.
	//
	// Required if `use_discovery` is false or omitted.
	TokenEndpoint string `yaml:"token_endpoint" json:"token_endpoint,omitempty" koanf:"token_endpoint"`
	// `use_discovery` determines if configuration information about an OpenID Connect (OIDC) provider, such as
	// endpoint URLs and supported features,should be automatically retrieved, from a well-known
	// URL (typically /.well-known/openid-configuration).
	UseDiscovery bool `yaml:"use_discovery" json:"use_discovery,omitempty" koanf:"use_discovery" jsonschema:"default=true"`
	// URL of the provider's endpoint that returns claims about an authenticated end-user.
	//
	// Required if `use_discovery` is false or omitted.
	UserinfoEndpoint string `yaml:"userinfo_endpoint" json:"userinfo_endpoint,omitempty" koanf:"userinfo_endpoint"`
}

func (p *CustomThirdPartyProvider) Validate() error {
	if p.Enabled {
		if p.DisplayName == "" {
			return errors.New("missing display_name")
		}
		if p.ClientID == "" {
			return errors.New("missing client_id")
		}
		if p.Secret == "" {
			return errors.New("missing client secret")
		}
		if len(p.Scopes) == 0 {
			return errors.New("missing scopes")
		}
		if p.UseDiscovery == true {
			if p.Issuer == "" {
				return errors.New("issuer must be set when use_discovery is set to true")
			}
		} else {
			authorizationEndpointSet := p.AuthorizationEndpoint != ""
			tokenEndpointSet := p.TokenEndpoint != ""
			userinfoEndpointSet := p.UserinfoEndpoint != ""

			if !authorizationEndpointSet || !tokenEndpointSet || !userinfoEndpointSet {
				return errors.New("authorization_endpoint, token_endpoint and userinfo_endpoint must be set when use_discovery is set to false or unset")
			}
		}
	}

	return nil
}

func (CustomThirdPartyProvider) JSONSchemaExtend(schema *jsonschema.Schema) {
	schema.Title = "custom_provider"

	enabledFalseOrNull := &jsonschema.Schema{Properties: orderedmap.New[string, *jsonschema.Schema]()}
	enabledFalseOrNull.Properties.Set("enabled", &jsonschema.Schema{
		AnyOf: []*jsonschema.Schema{
			{Const: false},
			{Type: "null"},
		},
	})

	useDiscoveryFalseOrNull := &jsonschema.Schema{Properties: orderedmap.New[string, *jsonschema.Schema]()}
	useDiscoveryFalseOrNull.Properties.Set("use_discovery", &jsonschema.Schema{
		AnyOf: []*jsonschema.Schema{
			{Const: false},
			{Type: "null"},
		},
	})

	endpointsRequired := &jsonschema.Schema{
		Required: []string{"authorization_endpoint", "token_endpoint", "userinfo_endpoint"},
	}

	issuerRequired := &jsonschema.Schema{
		Required: []string{"issuer"},
	}

	schema.If = enabledFalseOrNull
	schema.Then = &jsonschema.Schema{
		Required: []string{},
	}
	schema.Else = &jsonschema.Schema{
		Required: []string{"display_name", "client_id", "secret", "scopes"},
		If:       useDiscoveryFalseOrNull,
		Then:     endpointsRequired,
		Else:     issuerRequired,
	}
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
	//`facebook` contains the provider configuration for Facebook.
	Facebook ThirdPartyProvider `yaml:"facebook" json:"facebook,omitempty" koanf:"facebook"`
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
	// `prompt` specifies whether the Authorization Server prompts the End-User for reauthentication and consent.
	// Possible values are:
	// - login
	// - none
	// - consent
	// - select_account
	// Please note that not all providers support all values. Check the corresponding docs of the provider for supported values.
	Prompt string `yaml:"prompt" json:"prompt,omitempty" koanf:"prompt" jsonschema:"default=consent"`
	// `secret` is the client secret for the OAuth/OIDC client. Must be obtained from the provider.
	//
	// Required if the provider is `enabled`.
	Secret string `yaml:"secret" json:"secret,omitempty" koanf:"secret"`
	// `ID` is a unique name/slug/identifier for the provider. It is the lowercased key of the corresponding field
	// in ThirdPartyProviders. See also: CustomThirdPartyProvider.ID.
	ID string `jsonschema:"-" yaml:"-" json:"-" koanf:"-"`
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
		Required: []string{},
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
