package thirdparty

import (
	"context"
	"errors"
	"fmt"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/mitchellh/mapstructure"
	"github.com/teamhanko/hanko/backend/config"
	"golang.org/x/oauth2"
	"net/mail"
	"regexp"
)

const (
	MicrosoftAuthBase           = "https://login.microsoftonline.com/common"
	MicrosoftKeysEndpoint       = "https://login.microsoftonline.com/common/discovery/v2.0/keys"
	MicrosoftOAuthAuthEndpoint  = MicrosoftAuthBase + "/oauth2/v2.0/authorize"
	MicrosoftOAuthTokenEndpoint = MicrosoftAuthBase + "/oauth2/v2.0/token"
)

var DefaultScopes = []string{
	"openid",
	"profile",
	"email",
}

type microsoftProvider struct {
	*oauth2.Config
}

type MicrosoftUser struct {
	ID                string `json:"id"`
	Name              string `json:"displayName"`
	Email             string `json:"mail"`
	EmailVerified     bool   `json:"email_verified"`
	UserPrincipalName string `json:"user_principal_name"`
}

// NewMicrosoftProvider creates a Microsoft third party provider.
func NewMicrosoftProvider(config config.ThirdPartyProvider, redirectURL string) (OAuthProvider, error) {
	if !config.Enabled {
		return nil, errors.New("microsoft provider is disabled")
	}

	return &microsoftProvider{
		Config: &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.Secret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  MicrosoftOAuthAuthEndpoint,
				TokenURL: MicrosoftOAuthTokenEndpoint,
			},
			Scopes:      DefaultScopes,
			RedirectURL: redirectURL,
		},
	}, nil
}

func (p microsoftProvider) GetOAuthToken(code string) (*oauth2.Token, error) {
	return p.Exchange(context.Background(), code)
}

func (p microsoftProvider) GetUserData(token *oauth2.Token) (*UserData, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("id_token missing")
	}

	jwks, err := jwk.Fetch(context.Background(), MicrosoftKeysEndpoint)
	if err != nil {
		return nil, err
	}

	parsedIDToken, err := jwt.Parse(
		[]byte(rawIDToken),
		// JWKs of the JWKS (see 'MicrosoftKeysEndpoint') do not contain an 'alg' field. jws.WithKeySet expects this
		// field to be present per default, hence usage of the extra option jws.WithInferAlgorithmFromKey.
		// See the jwt.WithKeySet documentation.
		jwt.WithKeySet(jwks, jws.WithInferAlgorithmFromKey(true)),
		jwt.WithAudience(p.Config.ClientID),
		jwt.WithValidator(p.issuerValidator()),
	)

	if err != nil {
		return nil, fmt.Errorf("could not parse id token: %w", err)
	}

	if parsedIDToken == nil {
		return nil, errors.New("could not parse id token")
	}

	idTokenClaims, err := p.getIdTokenClaims(parsedIDToken.PrivateClaims())
	if err != nil {
		return nil, fmt.Errorf("could not extract claims from id token: %w", err)
	}

	if idTokenClaims == nil {
		return nil, errors.New("id token claims must not be nil")
	}

	var email *Email
	if idTokenClaims.UserPrincipalName != "" {
		// Should be an email address, sanity check just to make sure.
		if address, err := mail.ParseAddress(idTokenClaims.UserPrincipalName); err == nil {
			email = &Email{
				Email: address.Address,
				// Assume it is verified because it looks like UPN suffixes cannot be set to unverified domains.
				Verified: true,
				Primary:  true,
			}
		}
	} else {
		emailIsVerified, emailVerificationError := idTokenClaims.IsEmailVerified()

		if emailVerificationError != nil {
			return nil, emailVerificationError
		}

		if emailIsVerified {
			email = &Email{
				Email:    idTokenClaims.Email,
				Verified: true,
				Primary:  true,
			}
		} else {
			email = &Email{
				Email:    idTokenClaims.Email,
				Verified: false,
				Primary:  true,
			}
		}
	}

	if email == nil {
		return nil, errors.New("unable to find email with Microsoft provider")
	}

	data := &UserData{}
	data.Emails = append(data.Emails, *email)

	data.Metadata = &Claims{
		Issuer:            parsedIDToken.Issuer(),
		Subject:           parsedIDToken.Subject(),
		Name:              idTokenClaims.Name,
		PreferredUsername: idTokenClaims.PreferredUsername,
		Email:             email.Email,
		EmailVerified:     email.Verified,
	}

	return data, nil
}

func (p microsoftProvider) Name() string {
	return "microsoft"
}

func (p microsoftProvider) issuerValidator() jwt.ValidatorFunc {
	var microsoftIssuerRegexp = regexp.MustCompile("^https://login[.]microsoftonline[.]com/([^/]+)/v2[.]0/?$")
	validator := jwt.ValidatorFunc(func(_ context.Context, t jwt.Token) jwt.ValidationError {
		if !microsoftIssuerRegexp.MatchString(t.Issuer()) {
			return jwt.NewValidationError(fmt.Errorf(`%s is not a valid microsoft issuer`, t.Issuer()))
		}
		return nil
	})
	return validator
}

type microsoftIdTokenClaims struct {
	Email                              string `mapstructure:"email"`
	Name                               string `mapstructure:"name"`
	PreferredUsername                  string `mapstructure:"preferred_username"`
	UserPrincipalName                  string `mapstructure:"upn"`
	XMicrosoftEmailDomainOwnerVerified any    `mapstructure:"xms_edov"`
}

// IsEmailVerified checks if the email used is verified. Functionality mainly derived from Supabase's GoTrue fork
// See: https://github.com/supabase/gotrue/blob/master/internal/api/provider/oidc.go#L221
// See also: https://www.descope.com/blog/post/noauth
func (c *microsoftIdTokenClaims) IsEmailVerified() (bool, error) {
	address, err := mail.ParseAddress(c.Email)

	if err != nil {
		return false, fmt.Errorf("could not parse email from email claim: %w", err)
	}

	if address == nil {
		return false, errors.New("could not extract email from email claim")
	}

	emailVerified := false

	edov := c.XMicrosoftEmailDomainOwnerVerified

	// If xms_edov is not set, and an email is present or xms_edov is true,
	// only then is the email regarded as verified.
	// https://learn.microsoft.com/en-us/azure/active-directory/develop/migrate-off-email-claim-authorization#using-the-xms_edov-optional-claim-to-determine-email-verification-status-and-migrate-users
	if edov == nil {
		// An email is provided, but xms_edov is not -- probably not
		// configured, so we must assume the email is verified as Azure
		// will only send out a potentially unverified email address in
		// single-tenant apps (which we do not support - only the multi-tenant
		// + public account type).
		emailVerified = true
	} else {
		edovBool := false

		// Azure can't be trusted with how they encode the xms_edov
		// claim. Sometimes it's "xms_edov": "1", sometimes "xms_edov": true.
		switch v := edov.(type) {
		case bool:
			edovBool = v
		case string:
			edovBool = v == "1" || v == "true"
		default:
			edovBool = false
		}

		emailVerified = edovBool
	}

	return emailVerified, nil
}

func (p microsoftProvider) getIdTokenClaims(privateClaims map[string]interface{}) (*microsoftIdTokenClaims, error) {
	var claims microsoftIdTokenClaims
	err := mapstructure.Decode(privateClaims, &claims)
	if err != nil {
		return nil, fmt.Errorf("failed to decode claims: %w", err)
	}

	return &claims, nil
}
