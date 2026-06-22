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

	"github.com/teamhanko/hanko/backend/v2/config"
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
	GetUserData(*oauth2.Token) (*UserData, error)
	GetOAuthToken(string, ...oauth2.AuthCodeOption) (*oauth2.Token, error)
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

func makeRequest(token *oauth2.Token, config *oauth2.Config, url string, dst interface{}) error {
	client := config.Client(context.Background(), token)
	client.Timeout = time.Second * 10
	res, err := client.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	bodyBytes, _ := io.ReadAll(res.Body)
	defer res.Body.Close()
	res.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return errors.New(string(bodyBytes))
	}

	if err := json.NewDecoder(res.Body).Decode(dst); err != nil {
		return err
	}

	return nil
}
