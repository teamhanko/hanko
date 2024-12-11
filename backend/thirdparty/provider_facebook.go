package thirdparty

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/teamhanko/hanko/backend/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
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
	*oauth2.Config
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
		Config: &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.Secret,
			Endpoint:     facebook.Endpoint,
			Scopes:       DefaultFacebookScopes,
			RedirectURL:  redirectURL,
		},
	}, nil
}

func (f facebookProvider) GetOAuthToken(code string) (*oauth2.Token, error) {
	return f.Exchange(context.Background(), code)
}

func (f facebookProvider) GetUserData(token *oauth2.Token) (*UserData, error) {
	var fbUser FacebookUser
	client := f.Client(context.Background(), token)
	resp, err := client.Get(FacebookUserInfoEndpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&fbUser); err != nil {
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
	return "facebook"
}
