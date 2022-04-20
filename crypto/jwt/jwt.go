package jwt

import (
	"context"
	"errors"
	"fmt"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type Generator struct {
	signatureKey     jwk.Key
	verificationKeys []jwk.Key
}

func NewGenerator(signatureKey jwk.Key, verificationKeys []jwk.Key) (*Generator, error) {
	if signatureKey == nil {
		return nil, errors.New("no key for signing was provided")
	}
	if len(verificationKeys) == 0 {
		return nil, errors.New("no keys for verification were provided")
	}
	var vKeys []jwk.Key
	for _, key := range verificationKeys {
		pKey, err := jwk.PublicKeyOf(key)
		if err != nil {
			return nil, fmt.Errorf("failed to get public key: %w", err)
		}
		vKeys = append(vKeys, pKey)
	}
	return &Generator{
		signatureKey:     signatureKey,
		verificationKeys: vKeys,
	}, nil
}

func (g *Generator) Sign(token jwt.Token) ([]byte, error) {
	signed, err := jwt.Sign(token, jwt.WithKey(jwa.RS256, g.signatureKey))
	if err != nil {
		return nil, fmt.Errorf("failed to sign jwt: %w", err)
	}
	return signed, nil
}

func (g *Generator) Verify(signed []byte) (jwt.Token, error) {
	token, err := jwt.Parse(signed, jwt.WithKeyProvider(g))
	if err != nil {
		return nil, fmt.Errorf("failed to verify jwt: %w", err)
	}
	return token, nil
}

func (g *Generator) FetchKeys(ctx context.Context, sink jws.KeySink, sig *jws.Signature, msg *jws.Message) error {
	for _, key := range g.verificationKeys {
		sink.Key(jwa.RS256, key)
	}
	return nil
}
