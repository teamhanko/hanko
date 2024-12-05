package thirdparty

import (
	"context"
	"fmt"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/mitchellh/mapstructure"
	"github.com/teamhanko/hanko/backend/config"
	"golang.org/x/oauth2"
)

type customProvider struct {
	config       *config.CustomThirdPartyProvider
	oauthConfig  *oauth2.Config
	oidcProvider *oidc.Provider
}

func NewCustomThirdPartyProvider(config *config.CustomThirdPartyProvider, redirectURL string) (OAuthProvider, error) {
	if !config.Enabled {
		return nil, fmt.Errorf("provider %s is disabled", config.Name)
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
	return p.oauthConfig.AuthCodeURL(state, opts...)
}

func (p customProvider) GetOAuthToken(code string) (*oauth2.Token, error) {
	return p.oauthConfig.Exchange(context.Background(), code)
}

func (p customProvider) GetUserData(token *oauth2.Token) (*UserData, error) {
	tokenSource := p.oauthConfig.TokenSource(context.Background(), token)

	userInfo, err := p.oidcProvider.UserInfo(context.Background(), tokenSource)
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

	if p.config.AttributeMapping != nil {
		for hankoClaim, providerClaim := range p.config.AttributeMapping {
			userInfoClaims[hankoClaim] = userInfoClaims[providerClaim]
			delete(userInfoClaims, providerClaim)
		}
	}

	var claims Claims
	err = mapstructure.Decode(userInfoClaims, &claims)
	if err != nil {
		return nil, fmt.Errorf("could not get user data: %s", err)
	}

	if claims.Email == "" {
		return nil, fmt.Errorf("could not get user data: email not present")
	}

	return &UserData{
		Metadata: &claims,
	}, nil
}

func (p customProvider) Name() string {
	return p.config.Name
}
