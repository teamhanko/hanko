package thirdparty

import (
	"context"
	"errors"

	"github.com/teamhanko/hanko/backend/config"
	"golang.org/x/oauth2"
)

const (
	MicrosoftAuthBase  = "https://login.microsoftonline.com/common"
	MicrosoftGraphAPI  = "https://graph.microsoft.com/v1.0"
	OAuthEndpoint      = MicrosoftAuthBase + "/oauth2/v2.0/authorize"
	OAuthTokenEndpoint = MicrosoftAuthBase + "/oauth2/v2.0/token"
	UserInfoEndpoint   = MicrosoftGraphAPI + "/me"
)

var DefaultScopes = []string{
	"User.Read",
}

type microsoftProvider struct {
	*oauth2.Config
}

type MicrosoftUser struct {
	ID            string `json:"id"`
	Name          string `json:"displayName"`
	Email         string `json:"mail"`
	EmailVerified bool   `json:"email_verified"`
}

// NewMicrosoftProvider creates a Microsoft third party provider.
func NewMicrosoftProvider(config config.ThirdPartyProvider, redirectURL string) (OAuthProvider, error) {
	if !config.Enabled {
		return nil, errors.New("microsoft provider is disabled")
	}

	return &microsoftProvider{
		Config: &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.Secret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  OAuthEndpoint,
				TokenURL: OAuthTokenEndpoint,
			},
			Scopes:      DefaultScopes,
			RedirectURL: redirectURL,
		},
	}, nil
}

func (g microsoftProvider) GetOAuthToken(code string) (*oauth2.Token, error) {
	return g.Exchange(context.Background(), code)
}

func (g microsoftProvider) GetUserData(token *oauth2.Token) (*UserData, error) {
	var user MicrosoftUser
	if err := makeRequest(token, g.Config, UserInfoEndpoint, &user); err != nil {
		return nil, err
	}

	data := &UserData{}

	if user.Email != "" {
		data.Emails = append(data.Emails, Email{
			Email: user.Email,
			// Defaults to true.
			Verified: true,
			Primary:  true,
		})
	}

	if len(data.Emails) <= 0 {
		return nil, errors.New("unable to find email with Microsoft provider")
	}

	data.Metadata = &Claims{
		Issuer:  MicrosoftAuthBase,
		Subject: user.ID,
		Name:    user.Name,
		Email:   user.Email,
		// Defaults to true.
		EmailVerified: true,
	}

	return data, nil
}

func (g microsoftProvider) Name() string {
	return "microsoft"
}
