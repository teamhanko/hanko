package thirdparty

import (
	"context"
	"errors"
	"fmt"

	"github.com/teamhanko/hanko/backend/config"
	"golang.org/x/oauth2"
)

const (
	DiscordAPIBase            = "https://discord.com/api"
	DiscordOauthAuthEndpoint  = "https://discord.com/oauth2/authorize"
	DiscordOauthTokenEndpoint = DiscordAPIBase + "/oauth2/token"
	DiscordUserInfoEndpoint   = DiscordAPIBase + "/users/@me"
)

var DefaultDiscordScopes = []string{
	"identify",
	"email",
}

type discordProvider struct {
	*oauth2.Config
}

type DiscordUser struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	GlobalName string `json:"global_name"`
	Avatar     string `json:"avatar"`
	Email      string `json:"email"`
	Verified   bool   `json:"verified"`
}

// NewDiscordProvider creates a Discord third party provider.
func NewDiscordProvider(config config.ThirdPartyProvider, redirectURL string) (OAuthProvider, error) {
	if !config.Enabled {
		return nil, errors.New("discord provider is disabled")
	}

	return &discordProvider{
		Config: &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.Secret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  DiscordOauthAuthEndpoint,
				TokenURL: DiscordOauthTokenEndpoint,
			},
			Scopes:      DefaultDiscordScopes,
			RedirectURL: redirectURL,
		},
	}, nil
}

func (g discordProvider) GetOAuthToken(code string) (*oauth2.Token, error) {
	return g.Exchange(context.Background(), code)
}

func (g discordProvider) GetUserData(token *oauth2.Token) (*UserData, error) {
	var user DiscordUser
	if err := makeRequest(token, g.Config, DiscordUserInfoEndpoint, &user); err != nil {
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
		return nil, errors.New("unable to find email with Discord provider")
	}

	data.Metadata = &Claims{
		Issuer:        DiscordAPIBase,
		Subject:       user.ID,
		Name:          user.GlobalName,
		Picture:       g.buildAvatarURL(user.ID, user.Avatar),
		Email:         user.Email,
		EmailVerified: user.Verified,
	}

	return data, nil
}

func (g discordProvider) buildAvatarURL(userID string, avatarHash string) string {
	if avatarHash == "" {
		return "" // No image
	}

	return fmt.Sprintf("https://cdn.discordapp.com/avatars/%s/%s.png", userID, avatarHash)
}

func (g discordProvider) Name() string {
	return "discord"
}
