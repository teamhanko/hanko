package thirdparty

import (
	"context"
	"errors"
	"github.com/teamhanko/hanko/backend/config"
	"golang.org/x/oauth2"
)

const (
	FacebookAuthBase           = "https://www.facebook.com"
	FacebookAPIBase            = "https://graph.facebook.com"
	FacebookOauthAuthEndpoint  = FacebookAuthBase + "/v21.0/dialog/oauth"
	FacebookOauthTokenEndpoint = FacebookAPIBase + "/v21.0/oauth/access_token"
	FacebookUserInfoEndpoint   = FacebookAPIBase + "/me?fields=id,name,email,picture"
)

var DefaultFacebookScopes = []string{
	"email", "public_profile",
}

type facebookProvider struct {
	config      config.ThirdPartyProvider
	oauthConfig *oauth2.Config
}

type FacebookUser struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Picture struct {
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
	} `json:"picture"`
}

// NewFacebookProvider creates a Facebook third-party OAuth provider.
func NewFacebookProvider(config config.ThirdPartyProvider, redirectURL string) (OAuthProvider, error) {
	if !config.Enabled {
		return nil, errors.New("facebook provider is disabled")
	}

	return &facebookProvider{
		config: config,
		oauthConfig: &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.Secret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  FacebookOauthAuthEndpoint,
				TokenURL: FacebookOauthTokenEndpoint,
			},
			Scopes:      DefaultFacebookScopes,
			RedirectURL: redirectURL,
		},
	}, nil
}

func (f facebookProvider) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return f.oauthConfig.AuthCodeURL(state, opts...)
}

func (f facebookProvider) GetOAuthToken(code string) (*oauth2.Token, error) {
	return f.oauthConfig.Exchange(context.Background(), code)
}

func (f facebookProvider) GetUserData(token *oauth2.Token) (*UserData, error) {
	var fbUser FacebookUser
	if err := makeRequest(token, f.oauthConfig, FacebookUserInfoEndpoint, &fbUser); err != nil {
		return nil, err
	}

	data := &UserData{
		Emails: []Email{
			{
				Email:    fbUser.Email,
				Verified: true, // Facebook email is considered verified
				Primary:  true,
			},
		},
		Metadata: &Claims{
			Issuer:        FacebookAuthBase,
			Subject:       fbUser.ID,
			Name:          fbUser.Name,
			Picture:       fbUser.Picture.Data.URL,
			Email:         fbUser.Email,
			EmailVerified: true,
		},
	}

	return data, nil
}

func (f facebookProvider) Name() string {
	return f.config.Name
}
