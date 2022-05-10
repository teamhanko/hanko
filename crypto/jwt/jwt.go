package jwt

import (
	"errors"
	"fmt"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type Generator interface {
	Sign(jwt.Token) ([]byte, error)
	Verify([]byte) (jwt.Token, error)
}

// Generator is used to sign and verify JWTs
type generator struct {
	signatureKey jwk.Key
	verKeys      jwk.Set
}

// NewGenerator returns a new jwt generator which signs JWTs with the given signing key and verifies JWTs with the given verificationKeys
func NewGenerator(signatureKey jwk.Key, verificationKeys jwk.Set) (Generator, error) {
	if signatureKey == nil {
		return nil, errors.New("no key for signing was provided")
	}
	if verificationKeys.Len() == 0 {
		return nil, errors.New("no keys for verification were provided")
	}
	pubKeySet, err := jwk.PublicSetOf(verificationKeys)
	if err != nil {
		return nil, err
	}
	return &generator{
		signatureKey: signatureKey,
		verKeys:      pubKeySet,
	}, nil
}

// Sign a JWT with the signing key and returns it
func (g *generator) Sign(token jwt.Token) ([]byte, error) {
	signed, err := jwt.Sign(token, jwt.WithKey(jwa.RS256, g.signatureKey))
	if err != nil {
		return nil, fmt.Errorf("failed to sign jwt: %w", err)
	}
	return signed, nil
}

// Verify verifies a JWT, using the verificationKeys and returns the parsed JWT
func (g *generator) Verify(signed []byte) (jwt.Token, error) {
	token, err := jwt.Parse(signed, jwt.WithKeySet(g.verKeys))
	if err != nil {
		return nil, fmt.Errorf("failed to verify jwt: %w", err)
	}
	return token, nil
}
