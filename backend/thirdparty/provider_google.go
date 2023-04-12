package thirdparty

import (
	"context"
	"errors"
	"github.com/teamhanko/hanko/backend/config"
	"golang.org/x/oauth2"
)

const (
	GoogleAuthBase           = "accounts.google.com"
	GoogleAPIBase            = "www.googleapis.com"
	GoogleOauthAuthEndpoint  = "/o/oauth2/auth"
	GoogleOauthTokenEndpoint = "/o/oauth2/token"
	GoogleUserInfoEndpoint   = "/oauth2/v3/userinfo"
)

var DefaultGoogleScopes = []string{
	"email",
}

type googleProvider struct {
	*oauth2.Config
	APIPath string
}

type GoogleUser struct {
	ID            string `json:"sub"`
	Name          string `json:"name"`
	AvatarURL     string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

// NewGoogleProvider creates a Google third party provider.
func NewGoogleProvider(config config.ThirdPartyProvider, redirectURL string) (OAuthProvider, error) {
	if !config.Enabled {
		return nil, errors.New("google provider is disabled")
	}

	return &googleProvider{
		Config: &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.Secret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://" + GoogleAuthBase + GoogleOauthAuthEndpoint,
				TokenURL: "https://" + GoogleAuthBase + GoogleOauthTokenEndpoint,
			},
			Scopes:      DefaultGoogleScopes,
			RedirectURL: redirectURL,
		},
		APIPath: "https://" + GoogleAPIBase,
	}, nil
}

func (g googleProvider) GetOAuthToken(code string) (*oauth2.Token, error) {
	return g.Exchange(context.Background(), code)
}

func (g googleProvider) GetUserData(token *oauth2.Token) (*UserData, error) {
	var user GoogleUser
	if err := makeRequest(token, g.Config, g.APIPath+GoogleUserInfoEndpoint, &user); err != nil {
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

	data.Metadata = &Claims{
		Issuer:        g.APIPath,
		Subject:       user.ID,
		Name:          user.Name,
		Picture:       user.AvatarURL,
		Email:         user.Email,
		EmailVerified: user.EmailVerified,
	}

	return data, nil
}

func (g googleProvider) Name() string {
	return "google"
}
