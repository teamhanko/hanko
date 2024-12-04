package thirdparty

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/teamhanko/hanko/backend/config"
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

type ClaimsAddress struct {
	Formatted  string `json:"formatted,omitempty" mapstructure:"formatted,omitempty"`
	Locality   string `json:"locality,omitempty" mapstructure:"locality,omitempty"`
	PostalCode string `json:"postal_code,omitempty" mapstructure:"postal_code,omitempty"`
	Region     string `json:"region,omitempty" mapstructure:"region,omitempty"`
	Street     string `json:"street_address,omitempty" mapstructure:"street_address,omitempty"`
}

type Claims struct {
	// Reserved claims
	Issuer  string  `json:"iss,omitempty" mapstructure:"iss,omitempty"`
	Subject string  `json:"sub,omitempty" mapstructure:"sub,omitempty"`
	Aud     string  `json:"aud,omitempty" mapstructure:"aud,omitempty"`
	Iat     float64 `json:"iat,omitempty" mapstructure:"iat,omitempty"`
	Exp     float64 `json:"exp,omitempty" mapstructure:"exp,omitempty"`

	// Default profile claims
	Address           *ClaimsAddress `json:"address,omitempty" mapstructure:"address,omitempty"`
	Birthdate         string         `json:"birthdate,omitempty" mapstructure:"birthdate,omitempty"`
	Email             string         `json:"email,omitempty" mapstructure:"email,omitempty"`
	EmailVerified     bool           `json:"email_verified,omitempty" mapstructure:"email_verified,omitempty"`
	FamilyName        string         `json:"family_name,omitempty" mapstructure:"family_name,omitempty"`
	Gender            string         `json:"gender,omitempty" mapstructure:"gender,omitempty"`
	GivenName         string         `json:"given_name,omitempty" mapstructure:"given_name,omitempty"`
	Locale            string         `json:"locale,omitempty" mapstructure:"locale,omitempty"`
	MiddleName        string         `json:"middle_name,omitempty" mapstructure:"middle_name,omitempty"`
	Name              string         `json:"name,omitempty" mapstructure:"name,omitempty"`
	NickName          string         `json:"nickname,omitempty" mapstructure:"nickname,omitempty"`
	Phone             string         `json:"phone,omitempty" mapstructure:"phone,omitempty"`
	PhoneVerified     bool           `json:"phone_verified,omitempty" mapstructure:"phone_verified,omitempty"`
	Picture           string         `json:"picture,omitempty" mapstructure:"picture,omitempty"`
	PreferredUsername string         `json:"preferred_username,omitempty" mapstructure:"preferred_username,omitempty"`
	Profile           string         `json:"profile,omitempty" mapstructure:"profile,omitempty"`
	UpdatedAt         string         `json:"updated_at,omitempty" mapstructure:"updated_at,omitempty"`
	Website           string         `json:"website,omitempty" mapstructure:"website,omitempty"`
	ZoneInfo          string         `json:"zoneinfo,omitempty" mapstructure:"zoneinfo,omitempty"`

	// Custom profile claims that are oidc specific
	CustomClaims map[string]interface{} `json:"custom_claims,omitempty" mapstructure:"custom_claims,remain,omitempty"`
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
	GetOAuthToken(string) (*oauth2.Token, error)
	Name() string
}

func GetProvider(config config.ThirdParty, id string) (OAuthProvider, error) {
	idLower := strings.ToLower(id)

	if strings.HasPrefix(idLower, "custom_") {
		return getCustomThirdPartyProvider(config, idLower)
	} else {
		return getThirdPartyProvider(config, id)
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
