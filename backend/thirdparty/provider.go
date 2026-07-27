package thirdparty

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"

	"github.com/teamhanko/hanko/backend/v3/config"
	"golang.org/x/oauth2"
)

type UserData struct {
	Emails   Emails
	Metadata *Claims
}

func (u *UserData) ToMap() (map[string]interface{}, error) {
	data := make(map[string]interface{})
	if u.Metadata != nil {
		err := mapstructure.Decode(u.Metadata, &data)
		if err != nil {
			return nil, fmt.Errorf("could not convert user data to map: %w", err)
		}
	}
	return data, nil
}

type Emails []Email

type Email struct {
	Email    string
	Verified bool
	Primary  bool
}

type OAuthProvider interface {
	AuthCodeURL(string, ...oauth2.AuthCodeOption) string
	GetUserData(context.Context, *oauth2.Token) (*UserData, error)
	GetOAuthToken(context.Context, string, ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	ID() string
}

func GetProvider(config config.ThirdParty, id string) (OAuthProvider, error) {
	idLower := strings.ToLower(id)

	if strings.HasPrefix(idLower, "custom_") {
		return getCustomThirdPartyProvider(config, idLower)
	} else {
		return getThirdPartyProvider(config, idLower)
	}
}

func getCustomThirdPartyProvider(config config.ThirdParty, id string) (OAuthProvider, error) {
	if config.CustomProviders != nil {
		if providerConfig, ok := config.CustomProviders[strings.TrimPrefix(id, "custom_")]; ok {
			oauthProvider, err := NewCustomThirdPartyProvider(&providerConfig, config.RedirectURL)
			if err != nil {
				return nil, err
			}
			return oauthProvider, nil
		}
	}
	return nil, fmt.Errorf("unknown provider: %s", id)
}

func getThirdPartyProvider(config config.ThirdParty, id string) (OAuthProvider, error) {
	switch id {
	case "google":
		return NewGoogleProvider(config.Providers.Google, config.RedirectURL)
	case "github":
		return NewGithubProvider(config.Providers.GitHub, config.RedirectURL)
	case "apple":
		return NewAppleProvider(config.Providers.Apple, config.RedirectURL)
	case "discord":
		return NewDiscordProvider(config.Providers.Discord, config.RedirectURL)
	case "microsoft":
		return NewMicrosoftProvider(config.Providers.Microsoft, config.RedirectURL)
	case "linkedin":
		return NewLinkedInProvider(config.Providers.LinkedIn, config.RedirectURL)
	case "facebook":
		return NewFacebookProvider(config.Providers.Facebook, config.RedirectURL)
	default:
		return nil, fmt.Errorf("unknown provider: %s", id)
	}
}

func makeRequest(ctx context.Context, token *oauth2.Token, config *oauth2.Config, url string, dst interface{}) error {
	client := config.Client(ctx, token)
	client.Timeout = time.Second * 10
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return errors.New(string(bodyBytes))
	}

	if err := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(dst); err != nil {
		return err
	}

	return nil
}
