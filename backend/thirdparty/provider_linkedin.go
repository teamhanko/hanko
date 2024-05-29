package thirdparty

import (
	"context"
	"errors"
	"fmt"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/teamhanko/hanko/backend/config"
	"golang.org/x/oauth2"
)

const (
	LinkedInIssuer = "https://www.linkedin.com/oauth"
)

var DefaultLinkedinScopes = []string{
	"openid",
	"profile",
	"email",
}

type LinkedinUser struct {
	ID         string `json:"sub"`
	Name       string `json:"name"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Picture    string `json:"picture"`
	Locale     struct {
		Country  string `json:"country"`
		Language string `json:"language"`
	} `json:"locale"`
	Email    string `json:"email"`
	Verified bool   `json:"email_verified"`
}

type linkedInProvider struct {
	oidc *oidc.Provider
	*oauth2.Config
}

// NewLinkedInProvider creates a LinkedIn third party provider.
func NewLinkedInProvider(config config.ThirdPartyProvider, redirectURL string) (OAuthProvider, error) {
	if !config.Enabled {
		return nil, errors.New("linkedIn provider is disabled")
	}

	oidcProvider, err := oidc.NewProvider(context.Background(), LinkedInIssuer)
	if err != nil {
		return nil, err
	}
	endpoint := oidcProvider.Endpoint()

	return &linkedInProvider{
		oidc: oidcProvider,
		Config: &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.Secret,
			Endpoint:     endpoint,
			Scopes:       DefaultLinkedinScopes,
			RedirectURL:  redirectURL,
		},
	}, nil
}

func (g linkedInProvider) GetOAuthToken(code string) (*oauth2.Token, error) {
	return g.Exchange(context.Background(), code)
}

func (g linkedInProvider) GetUserData(token *oauth2.Token) (*UserData, error) {
	var user LinkedinUser
	if err := makeRequest(token, g.Config, g.oidc.UserInfoEndpoint(), &user); err != nil {
		return nil, err
	}

	data := &UserData{}

	if user.Email != "" {
		data.Emails = append(data.Emails, Email{
			Email:    user.Email,
			Verified: user.Verified,
			Primary:  true,
		})
	}

	if len(data.Emails) <= 0 {
		return nil, errors.New("unable to find email with LinkedIn provider")
	}

	data.Metadata = &Claims{
		Issuer:        LinkedInIssuer,
		Subject:       user.ID,
		Name:          user.Name,
		FamilyName:    user.FamilyName,
		GivenName:     user.GivenName,
		Picture:       user.Picture,
		Locale:        fmt.Sprintf("%s-%s", user.Locale.Country, user.Locale.Language),
		Email:         user.Email,
		EmailVerified: user.Verified,
	}

	return data, nil
}

func (g linkedInProvider) Name() string {
	return "linkedin"
}
