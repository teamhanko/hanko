package thirdparty

import (
	"context"
	"errors"
	"github.com/teamhanko/hanko/backend/config"
	"golang.org/x/oauth2"
	"strconv"
)

const (
	GithubAuthBase           = "github.com"
	GithubAPIBase            = "api.github.com"
	GithubOauthAuthEndpoint  = "/login/oauth/authorize"
	GithubOauthTokenEndpoint = "/login/oauth/access_token"
	GithubUserInfoEndpoint   = "/user"
	GitHubEmailsEndpoint     = "/user/emails"
)

var DefaultGitHubScopes = []string{
	"user:email",
}

type githubProvider struct {
	*oauth2.Config
	APIPath string
}

type githubUser struct {
	ID        int    `json:"id"`
	UserName  string `json:"login"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

type githubUserEmail struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

func NewGithubProvider(config config.ThirdPartyProvider, redirectURL string) (OAuthProvider, error) {
	if !config.Enabled {
		return nil, errors.New("github provider requested but disabled")
	}

	return &githubProvider{
		Config: &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.Secret,
			Endpoint: oauth2.Endpoint{
				// https://docs.github.com/en/developers/apps/building-oauth-apps/authorizing-oauth-apps#1-request-a-users-github-identity
				AuthURL: "https://" + GithubAuthBase + GithubOauthAuthEndpoint,
				// https://docs.github.com/en/developers/apps/building-oauth-apps/authorizing-oauth-apps#2-users-are-redirected-back-to-your-site-by-github
				TokenURL: "https://" + GithubAuthBase + GithubOauthTokenEndpoint,
			},
			RedirectURL: redirectURL,
			Scopes:      DefaultGitHubScopes,
		},
		APIPath: "https://" + GithubAPIBase,
	}, nil
}

func (g githubProvider) GetOAuthToken(code string) (*oauth2.Token, error) {
	return g.Exchange(context.Background(), code)
}

func (g githubProvider) GetUserData(token *oauth2.Token) (*UserData, error) {
	var user githubUser

	// https://docs.github.com/en/rest/users/users?apiVersion=2022-11-28#get-the-authenticated-user
	if err := makeRequest(token, g.Config, g.APIPath+GithubUserInfoEndpoint, &user); err != nil {
		return nil, err
	}

	data := &UserData{
		Metadata: &Claims{
			Issuer:            g.APIPath,
			Subject:           strconv.Itoa(user.ID),
			Name:              user.Name,
			Picture:           user.AvatarURL,
			PreferredUsername: user.UserName,
		},
	}

	var emails []*githubUserEmail
	// The user data 'email' value is the user's publicly visible email address. It is possible that the user
	// chose to not make this email public, hence the dedicated call to the 'emails' endpoint.
	// https://docs.github.com/en/rest/users/emails?apiVersion=2022-11-28#list-email-addresses-for-the-authenticated-user
	if err := makeRequest(token, g.Config, g.APIPath+GitHubEmailsEndpoint, &emails); err != nil {
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

func (g githubProvider) Name() string {
	return "github"
}
