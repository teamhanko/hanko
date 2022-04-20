package jwt

import (
	"fmt"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type Generator struct {
	privateKey *jwk.Key
	publicKey  *jwk.Key
}

func NewGenerator(privateKey *jwk.Key) (*Generator, error) {
	publicKey, err := jwk.PublicKeyOf(*privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create jwt generator: %w", err)
	}
	return &Generator{
		privateKey: privateKey,
		publicKey:  &publicKey,
	}, nil
}

func (g *Generator) Sign(token jwt.Token) ([]byte, error) {
	signed, err := jwt.Sign(token, jwt.WithKey(jwa.RS256, *g.privateKey))
	if err != nil {
		return nil, fmt.Errorf("failed to sign jwt: %w", err)
	}
	return signed, nil
}

func (g *Generator) Verify(signed []byte) (jwt.Token, error) {
	token, err := jwt.Parse(signed, jwt.WithKey(jwa.RS256, *g.publicKey))
	if err != nil {
		return nil, fmt.Errorf("failed to verify jwt: %w", err)
	}
	return token, nil
}
