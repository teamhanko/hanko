package thirdparty

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/mitchellh/mapstructure"
	zeroLogger "github.com/rs/zerolog/log"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/tidwall/gjson"
	"golang.org/x/oauth2"
)

type customProvider struct {
	config       *config.CustomThirdPartyProvider
	oauthConfig  *oauth2.Config
	oidcProvider *oidc.Provider
}

func NewCustomThirdPartyProvider(config *config.CustomThirdPartyProvider, redirectURL string) (OAuthProvider, error) {
	if !config.Enabled {
		return nil, fmt.Errorf("provider %s is disabled", config.ID)
	}

	customProvider := &customProvider{
		oauthConfig: &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.Secret,
			Scopes:       config.Scopes,
			RedirectURL:  redirectURL,
		},
	}

	if config.UseDiscovery {
		provider, err := oidc.NewProvider(context.Background(), config.Issuer)
		if err != nil {
			return nil, err
		}

		customProvider.oidcProvider = provider
		customProvider.oauthConfig.Endpoint = customProvider.oidcProvider.Endpoint()
	} else {
		providerConfig := oidc.ProviderConfig{
			IssuerURL:   config.Issuer,
			AuthURL:     config.AuthorizationEndpoint,
			TokenURL:    config.TokenEndpoint,
			UserInfoURL: config.UserinfoEndpoint,
			// Algorithms:  []string{"RS256"}, // TODO: What should be the value for this?
		}

		customProvider.oidcProvider = providerConfig.NewProvider(context.Background())
		customProvider.oauthConfig.Endpoint = customProvider.oidcProvider.Endpoint()
	}

	customProvider.config = config
	return customProvider, nil
}

func (p customProvider) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {

	if prompt := p.config.Prompt; prompt != "" {
		opts = append(opts, oauth2.SetAuthURLParam("prompt", prompt))
	}
	if acrValues := p.config.AcrValues; len(acrValues) > 0 {
		opts = append(opts, oauth2.SetAuthURLParam("acr_values", strings.Join(acrValues, " ")))
	}

	return p.oauthConfig.AuthCodeURL(state, opts...)
}

func (p customProvider) GetOAuthToken(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return p.oauthConfig.Exchange(ctx, code, opts...)
}

func (p customProvider) GetUserData(ctx context.Context, token *oauth2.Token) (*UserData, error) {
	tokenSource := p.oauthConfig.TokenSource(ctx, token)

	userInfo, err := p.oidcProvider.UserInfo(ctx, tokenSource)
	if err != nil {
		return nil, err
	}

	// oidc.UserInfo does not make raw claims map publicly accessible,
	// hence the additional unmarshal via oidc.UserInfo.Claims method
	userInfoClaims := make(map[string]interface{})
	err = userInfo.Claims(&userInfoClaims)
	if err != nil {
		return nil, fmt.Errorf("could not get user data: %s", err)
	}

	// Resolve CustomClaimMapping against the raw claims before AttributeMapping below
	// renames any of them - the two mapping mechanisms must read from the same untouched
	// source, or a claim AttributeMapping already consumed would be gone by the time
	// custom-claim resolution looks for it (see CustomClaimMapping's doc comment).
	customClaimSource := buildCustomClaimSource(p.config.CustomClaimMapping, userInfoClaims)

	if p.config.AttributeMapping != nil {
		for hankoClaim, providerClaim := range p.config.AttributeMapping {
			userInfoClaims[hankoClaim] = userInfoClaims[providerClaim]
		}
	}

	var claims Claims
	err = mapstructure.Decode(userInfoClaims, &claims)
	if err != nil {
		return nil, fmt.Errorf("could not get user data: %s", err)
	}

	return &UserData{
		Metadata:          &claims,
		CustomClaimSource: customClaimSource,
	}, nil
}

// buildCustomClaimSource resolves each CustomClaimMapping value - a plain top-level claim
// name or a gjson path (https://github.com/tidwall/gjson#path-syntax) - against the raw,
// pre-AttributeMapping OIDC claims. Attributes is keyed by the same path string used in
// mapping, since that's simpler than re-keying by hanko claim name and ResolveCustomClaims
// only ever looks a value up via mapping's value anyway.
func buildCustomClaimSource(mapping map[string]string, rawClaims map[string]interface{}) *CustomClaimSource {
	if len(mapping) == 0 {
		return nil
	}

	rawClaimsJSON, err := json.Marshal(rawClaims)
	if err != nil {
		zeroLogger.Warn().
			Err(err).
			Str("component", "thirdparty").
			Str("operation", "build_custom_claim_source").
			Msg("could not marshal raw OIDC claims for custom claim resolution")
		return nil
	}

	attributes := make(map[string]any, len(mapping))
	for _, path := range mapping {
		if result := gjson.GetBytes(rawClaimsJSON, path); result.Exists() {
			attributes[path] = result.Value()
		}
	}

	return &CustomClaimSource{
		Mapping:    mapping,
		Attributes: attributes,
	}
}

func (p customProvider) ID() string {
	return p.config.ID
}
