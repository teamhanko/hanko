package thirdparty

import (
	"context"
	"errors"
	"github.com/teamhanko/hanko/backend/v2/config"
	"golang.org/x/oauth2"
	"strconv"
)

const (
	GithubAuthBase           = "https://github.com"
	GithubAPIBase            = "https://api.github.com"
	GithubOauthAuthEndpoint  = GithubAuthBase + "/login/oauth/authorize"
	GithubOauthTokenEndpoint = GithubAuthBase + "/login/oauth/access_token"
	GithubUserInfoEndpoint   = GithubAPIBase + "/user"
	GitHubEmailsEndpoint     = GithubAPIBase + "/user/emails"
)

var DefaultGitHubScopes = []string{
	"user:email",
}

type githubProvider struct {
	config      config.ThirdPartyProvider
	oauthConfig *oauth2.Config
}

type GithubUser struct {
	ID        int    `json:"id"`
	UserName  string `json:"login"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

type GithubUserEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

func NewGithubProvider(config config.ThirdPartyProvider, redirectURL string) (OAuthProvider, error) {
	if !config.Enabled {
		return nil, errors.New("github provider requested but disabled")
	}

	return &githubProvider{
		config: config,
		oauthConfig: &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.Secret,
			Endpoint: oauth2.Endpoint{
				// https://docs.github.com/en/developers/apps/building-oauth-apps/authorizing-oauth-apps#1-request-a-users-github-identity
				AuthURL: GithubOauthAuthEndpoint,
				// https://docs.github.com/en/developers/apps/building-oauth-apps/authorizing-oauth-apps#2-users-are-redirected-back-to-your-site-by-github
				TokenURL: GithubOauthTokenEndpoint,
			},
			RedirectURL: redirectURL,
			Scopes:      DefaultGitHubScopes,
		},
	}, nil
}

func (g githubProvider) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return g.oauthConfig.AuthCodeURL(state, opts...)
}

func (g githubProvider) GetOAuthToken(code string) (*oauth2.Token, error) {
	return g.oauthConfig.Exchange(context.Background(), code)
}

func (g githubProvider) GetUserData(token *oauth2.Token) (*UserData, error) {
	var user GithubUser

	// https://docs.github.com/en/rest/users/users?apiVersion=2022-11-28#get-the-authenticated-user
	if err := makeRequest(token, g.oauthConfig, GithubUserInfoEndpoint, &user); err != nil {
		return nil, err
	}

	data := &UserData{
		Metadata: &Claims{
			Issuer:            GithubAuthBase,
			Subject:           strconv.Itoa(user.ID),
			Name:              user.Name,
			Picture:           user.AvatarURL,
			PreferredUsername: user.UserName,
		},
	}

	var emails []*GithubUserEmail
	// The user data 'email' value is the user's publicly visible email address. It is possible that the user
	// chose to not make this email public, hence the dedicated call to the 'emails' endpoint.
	// https://docs.github.com/en/rest/users/emails?apiVersion=2022-11-28#list-email-addresses-for-the-authenticated-user
	if err := makeRequest(token, g.oauthConfig, GitHubEmailsEndpoint, &emails); err != nil {
		return nil, err
	}

	for _, e := range emails {
		if e.Email != "" {
			data.Emails = append(data.Emails, Email{Email: e.Email, Verified: e.Verified, Primary: e.Primary})
		}

		if e.Primary {
			data.Metadata.Email = e.Email
			data.Metadata.EmailVerified = e.Verified
		}
	}

	if len(data.Emails) <= 0 {
		return nil, errors.New("unable to find email with GitHub provider")
	}

	return data, nil
}

func (g githubProvider) ID() string {
	return g.config.ID
}
