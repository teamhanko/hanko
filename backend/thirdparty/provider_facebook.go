package thirdparty

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/teamhanko/hanko/backend/v2/config"
	"golang.org/x/oauth2"
	"net/url"
)

const (
	FacebookAuthBase           = "https://www.facebook.com"
	FacebookAPIBase            = "https://graph.facebook.com"
	FacebookOauthAuthEndpoint  = FacebookAuthBase + "/v21.0/dialog/oauth"
	FacebookOauthTokenEndpoint = FacebookAPIBase + "/v21.0/oauth/access_token"
	FacebookUserInfoEndpoint   = FacebookAPIBase + "/me"
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
	FirstName  string `json:"first_name"`
	MiddleName string `json:"middle_name"`
	LastName   string `json:"last_name"`
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
	endpointURL, err := url.Parse(FacebookUserInfoEndpoint)
	if err != nil {
		return nil, err
	}

	endpointURLQuery := endpointURL.Query()
	endpointURLQuery.Add("fields", "id,name,email,picture,first_name,middle_name,last_name")

	// Calculate appsecret_proof, see:
	// https://developers.facebook.com/docs/graph-api/guides/secure-requests/#appsecret_proof
	hash := hmac.New(sha256.New, []byte(f.config.Secret))
	hash.Write([]byte(token.AccessToken))
	appsecretProof := hex.EncodeToString(hash.Sum(nil))

	endpointURLQuery.Add("appsecret_proof", appsecretProof)
	endpointURL.RawQuery = endpointURLQuery.Encode()

	var fbUser FacebookUser
	if err = makeRequest(token, f.oauthConfig, endpointURL.String(), &fbUser); err != nil {
		return nil, err
	}

	if fbUser.Email == "" {
		return nil, errors.New("unable to find email with Facebook provider")
	}

	data := &UserData{
		Emails: []Email{
			{
				Email: fbUser.Email,
				// Consider the email as verified because a User node only returns an email if a valid
				// email address is available. See: https://developers.facebook.com/docs/graph-api/reference/user/
				Verified: true,
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
			GivenName:     fbUser.FirstName,
			MiddleName:    fbUser.MiddleName,
			FamilyName:    fbUser.LastName,
		},
	}

	return data, nil
}

func (f facebookProvider) ID() string {
	return f.config.ID
}
