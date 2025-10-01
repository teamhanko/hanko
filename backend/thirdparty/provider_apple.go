package thirdparty

import (
	"context"
	"errors"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	zeroLogger "github.com/rs/zerolog/log"
	"github.com/teamhanko/hanko/backend/v2/config"
	"golang.org/x/oauth2"
	"net/url"
	"strconv"
	"strings"
)

const (
	AppleAPIBase       = "https://appleid.apple.com"
	AppleAuthEndpoint  = AppleAPIBase + "/auth/authorize"
	AppleTokenEndpoint = AppleAPIBase + "/auth/token"
	AppleKeysEndpoint  = AppleAPIBase + "/auth/keys"
)

var DefaultAppleScopes = []string{
	"name",
	"email",
}

type appleProvider struct {
	config      config.ThirdPartyProvider
	oauthConfig *oauth2.Config
}

func NewAppleProvider(config config.ThirdPartyProvider, redirectURL string) (OAuthProvider, error) {
	if !config.Enabled {
		return nil, errors.New("apple provider requested but disabled")
	}

	return &appleProvider{
		config: config,
		oauthConfig: &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.Secret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  AppleAuthEndpoint,
				TokenURL: AppleTokenEndpoint,
			},
			RedirectURL: redirectURL,
			Scopes:      DefaultAppleScopes,
		},
	}, nil
}

func (a appleProvider) AuthCodeURL(state string, args ...oauth2.AuthCodeOption) string {
	opts := append(args, oauth2.SetAuthURLParam("response_mode", "form_post"))
	authURL := a.oauthConfig.AuthCodeURL(state, opts...)
	u, _ := url.Parse(authURL)
	u.RawQuery = strings.ReplaceAll(u.RawQuery, "+", "%20")
	authURL = u.String()
	return authURL
}

func (a appleProvider) GetOAuthToken(code string) (*oauth2.Token, error) {
	return a.oauthConfig.Exchange(context.Background(), code)
}

func (a appleProvider) GetUserData(token *oauth2.Token) (*UserData, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("id_token missing")
	}

	set, err := jwk.Fetch(context.Background(), AppleKeysEndpoint)
	if err != nil {
		return nil, err
	}

	parsedIDToken, err := jwt.Parse(
		[]byte(rawIDToken),
		jwt.WithKeySet(set),
		jwt.WithIssuer(AppleAPIBase),
		jwt.WithAudience(a.oauthConfig.ClientID),
	)

	if err != nil {
		return nil, err
	}

	email, ok := parsedIDToken.PrivateClaims()["email"].(string)
	if !ok {
		return nil, errors.New("email claim expected to be of type string")
	}

	var emailVerified = false
	emailVerifiedRaw, found := parsedIDToken.PrivateClaims()["email_verified"]
	if found {
		switch v := emailVerifiedRaw.(type) {
		case string:
			emailVerified, err = strconv.ParseBool(v)
			if err != nil {
				zeroLogger.Warn().Err(err).Msgf("could not parse 'email_verified' claim as bool")
			}
		case bool:
			emailVerified = v
		default:
			zeroLogger.Warn().Msgf("'email_verified' claim is neither of type 'string' or 'bool'")
		}
	}

	userData := &UserData{
		Emails: []Email{{
			Email:    email,
			Verified: emailVerified,
			Primary:  true,
		}},
		Metadata: &Claims{
			Issuer:        parsedIDToken.Issuer(),
			Subject:       parsedIDToken.Subject(),
			Email:         email,
			EmailVerified: emailVerified,
		},
	}

	return userData, nil
}

func (a appleProvider) ID() string {
	return a.config.ID
}

func (a appleProvider) GetPromptParam() string {
	if a.config.Prompt != "" {
		return a.config.Prompt
	}
	return "consent"
}
