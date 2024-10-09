package thirdparty

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/teamhanko/hanko/backend/config"
	"golang.org/x/oauth2"
)

const (
	FacebookAPIBase            = "https://graph.facebook.com"
	FacebookOauthAuthEndpoint  = "https://www.facebook.com/v10.0/dialog/oauth"
	FacebookOauthTokenEndpoint = FacebookAPIBase + "/v10.0/oauth/access_token"
	FacebookUserInfoEndpoint   = FacebookAPIBase + "/me?fields=id,name,email,picture"
	FacebookDebugTokenEndpoint = FacebookAPIBase + "/debug_token"
)

var DefaultFacebookScopes = []string{
	"email",
	"public_profile",
}

type facebookProvider struct {
	*oauth2.Config
}

type FacebookUser struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Picture struct {
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
	} `json:"picture"`
	Email    string `json:"email"`
	Verified bool   `json:"verified"`
}

// TokenDebugResponse represents the response from the Facebook debug token API.
type TokenDebugResponse struct {
	Data struct {
		IsValid       bool   `json:"is_valid"`
		UserID        string `json:"user_id"`
		AppID         string `json:"app_id"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	} `json:"data"`
}

// NewFacebookProvider creates a Facebook third party provider.
func NewFacebookProvider(config config.ThirdPartyProvider, redirectURL string) (OAuthProvider, error) {
	if !config.Enabled {
		return nil, errors.New("facebook provider is disabled")
	}

	return &facebookProvider{
		Config: &oauth2.Config{
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

func (g facebookProvider) GetOAuthToken(code string) (*oauth2.Token, error) {
	return g.Exchange(context.Background(), code)
}

func (g facebookProvider) GetUserData(token *oauth2.Token) (*UserData, error) {
	client := g.Client(context.Background(), token)

	// Get user data from Facebook API.
	resp, err := client.Get(FacebookUserInfoEndpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user FacebookUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	// Verify email by calling the token debug endpoint
	emailVerified, err := g.verifyEmail(token)
	if err != nil {
		return nil, err
	}

	user.Verified = emailVerified

	data := &UserData{}
	if user.Email != "" {
		data.Emails = append(data.Emails, Email{
			Email:    user.Email,
			Verified: user.Verified,
			Primary:  true,
		})
	}

	if len(data.Emails) <= 0 {
		return nil, errors.New("unable to find email with Facebook provider")
	}

	data.Metadata = &Claims{
		Issuer:        FacebookAPIBase,
		Subject:       user.ID,
		Name:          user.Name,
		Picture:       user.Picture.Data.URL,
		Email:         user.Email,
		EmailVerified: user.Verified,
	}

	return data, nil
}

// verifyEmail checks if the user's email is verified using the Facebook debug token endpoint.
func (g facebookProvider) verifyEmail(token *oauth2.Token) (bool, error) {
	// Build the URL for the token debug endpoint
	appAccessToken := fmt.Sprintf("%s|%s", g.ClientID, g.ClientSecret)
	debugTokenURL := fmt.Sprintf("%s?input_token=%s&access_token=%s", FacebookDebugTokenEndpoint, token.AccessToken, appAccessToken)

	// Make the HTTP request to the Facebook token debug endpoint
	resp, err := http.Get(debugTokenURL)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var debugResponse TokenDebugResponse
	if err := json.NewDecoder(resp.Body).Decode(&debugResponse); err != nil {
		return false, err
	}

	// Check if the token is valid and if the email is verified
	if !debugResponse.Data.IsValid {
		return false, errors.New("invalid Facebook OAuth token")
	}

	return debugResponse.Data.EmailVerified, nil
}

func (g facebookProvider) Name() string {
	return "facebook"
}
