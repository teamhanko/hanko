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

	"github.com/fatih/structs"
	"github.com/teamhanko/hanko/backend/config"
	"golang.org/x/oauth2"
)

type UserData struct {
	Emails   Emails
	Metadata *Claims
}

func (u *UserData) ToMap() map[string]interface{} {
	var data map[string]interface{}
	if u.Metadata != nil {
		data = structs.Map(u.Metadata)
	}
	return data
}

type Claims struct {
	// Reserved claims
	Issuer  string  `json:"iss,omitempty" structs:"iss,omitempty"`
	Subject string  `json:"sub,omitempty" structs:"sub,omitempty"`
	Aud     string  `json:"aud,omitempty" structs:"aud,omitempty"`
	Iat     float64 `json:"iat,omitempty" structs:"iat,omitempty"`
	Exp     float64 `json:"exp,omitempty" structs:"exp,omitempty"`

	// Default profile claims
	Name              string `json:"name,omitempty" structs:"name,omitempty"`
	FamilyName        string `json:"family_name,omitempty" structs:"family_name,omitempty"`
	GivenName         string `json:"given_name,omitempty" structs:"given_name,omitempty"`
	MiddleName        string `json:"middle_name,omitempty" structs:"middle_name,omitempty"`
	NickName          string `json:"nickname,omitempty" structs:"nickname,omitempty"`
	PreferredUsername string `json:"preferred_username,omitempty" structs:"preferred_username,omitempty"`
	Profile           string `json:"profile,omitempty" structs:"profile,omitempty"`
	Picture           string `json:"picture,omitempty" structs:"picture,omitempty"`
	Website           string `json:"website,omitempty" structs:"website,omitempty"`
	Gender            string `json:"gender,omitempty" structs:"gender,omitempty"`
	Birthdate         string `json:"birthdate,omitempty" structs:"birthdate,omitempty"`
	ZoneInfo          string `json:"zoneinfo,omitempty" structs:"zoneinfo,omitempty"`
	Locale            string `json:"locale,omitempty" structs:"locale,omitempty"`
	UpdatedAt         string `json:"updated_at,omitempty" structs:"updated_at,omitempty"`
	Email             string `json:"email,omitempty" structs:"email,omitempty"`
	EmailVerified     bool   `json:"email_verified,omitempty" structs:"email_verified,omitempty"`
	Phone             string `json:"phone,omitempty" structs:"phone,omitempty"`
	PhoneVerified     bool   `json:"phone_verified,omitempty" structs:"phone_verified,omitempty"`

	// Custom profile claims that are provider specific
	CustomClaims map[string]interface{} `json:"custom_claims,omitempty" structs:"custom_claims,omitempty"`
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

func GetProvider(config config.ThirdParty, name string) (OAuthProvider, error) {
	n := strings.ToLower(name)

	switch n {
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
		return nil, fmt.Errorf("provider '%s' is not supported", name)
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
