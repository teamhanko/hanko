package thirdparty

import (
	"context"
	"errors"
	"fmt"
	"strings"

	oidc "github.com/coreos/go-oidc/v3/oidc"
	config "github.com/teamhanko/hanko/backend/config"
	oauth2 "golang.org/x/oauth2"
)

type genericOIDCProvider struct {
	*oauth2.Config
	oicProviderConfig *config.GenericOIDCProvider
	oidcProvider      *oidc.Provider
}

type GenericOAuth2User struct {
	ID            string `json:"sub"`
	Name          string `json:"name"`
	AvatarURL     string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

// NewGenericOIDCProvider creates a generic OIDC third party provider.
func NewGenericOIDCProvider(oicProviderConfig *config.GenericOIDCProvider, redirectURL string) (OAuthProvider, error) {
	if !oicProviderConfig.Enabled {
		return nil, fmt.Errorf("%s provider is disabled", oicProviderConfig.Slug)
	}
	oidcProvider, err := oidc.NewProvider(context.Background(), oicProviderConfig.Authority)
	if err != nil {
		return nil, err
	}
	endpoint := oidcProvider.Endpoint()
	scopes := strings.Split(oicProviderConfig.Scopes, " ")

	return &genericOIDCProvider{
		oicProviderConfig: oicProviderConfig,
		oidcProvider:      oidcProvider,
		Config: &oauth2.Config{
			ClientID:     oicProviderConfig.ClientID,
			ClientSecret: oicProviderConfig.Secret,
			Endpoint:     endpoint,
			Scopes:       scopes,
			RedirectURL:  redirectURL,
		},
	}, nil
}

func (g *genericOIDCProvider) GetOAuthToken(code string) (*oauth2.Token, error) {
	return g.Exchange(context.Background(), code)
}

func (g *genericOIDCProvider) GetUserData(token *oauth2.Token) (*UserData, error) {
	var user GenericOAuth2User
	if err := makeRequest(token, g.Config, g.oidcProvider.UserInfoEndpoint(), &user); err != nil {
		return nil, err
	}

	data := &UserData{}

	if user.Email != "" {
		data.Emails = append(data.Emails, Email{
			Email:    user.Email,
			Verified: user.EmailVerified,
			Primary:  true,
		})
	}

	if len(data.Emails) <= 0 {
		return nil, errors.New("unable to find email with Google provider")
	}
	emailVerified := user.EmailVerified
	if !g.oicProviderConfig.RequireProviderEmailVerification {
		// not required by the config, so we assume it's verified
		emailVerified = true
	}
	data.Metadata = &Claims{
		Issuer:        GoogleAuthBase,
		Subject:       user.ID,
		Name:          user.Name,
		Picture:       user.AvatarURL,
		Email:         user.Email,
		EmailVerified: emailVerified,
	}

	return data, nil
}

func (g *genericOIDCProvider) Name() string {
	return g.oicProviderConfig.Slug
}
