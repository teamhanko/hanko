package thirdparty

import (
	"errors"
	"fmt"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/crypto/jwk"
	"github.com/teamhanko/hanko/backend/session"
	"strings"
	"time"
)

func GenerateState(config config.ThirdParty, jwkManager jwk.Manager, provider string, redirectTo string) ([]byte, error) {
	if provider == "" {
		return nil, errors.New("provider must be present")
	}

	if redirectTo == "" {
		redirectTo = config.ErrorRedirectURL
	}

	now := time.Now().UTC()
	state, err := jwt.NewBuilder().
		IssuedAt(now).
		Expiration(now.Add(time.Minute*5)).
		Claim("provider", strings.ToLower(strings.TrimSpace(provider))).
		Claim("redirect_to", strings.TrimSpace(redirectTo)).
		Build()
	if err != nil {
		return nil, fmt.Errorf("could not generate token: %s", err)
	}
	signingKey, err := jwkManager.GetSigningKey()
	if err != nil {
		return nil, fmt.Errorf("could not get signing key: %s", err)
	}
	signedState, err := jwt.Sign(state, jwt.WithKey(jwa.RS256, signingKey))
	if err != nil {
		return nil, fmt.Errorf("could not sign token: %s", err)
	}

	return signedState, nil
}

type State struct {
	Provider   string
	RedirectTo string
}

func VerifyState(sessionManager session.Manager, state string) (*State, error) {
	verifiedToken, err := sessionManager.Verify(state)
	if err != nil {
		return nil, err
	}

	provider := verifiedToken.PrivateClaims()["provider"].(string)
	if provider == "" {
		return nil, errors.New("provider missing from state")
	}

	redirectTo := verifiedToken.PrivateClaims()["redirect_to"].(string)

	return &State{
		Provider:   provider,
		RedirectTo: redirectTo,
	}, nil
}
