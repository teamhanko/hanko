package jwk

import (
	"crypto/rand"
	"crypto/rsa"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

// RSAKeyGenerator
type RSAKeyGenerator struct {
}

func (g *RSAKeyGenerator) Generate(id string) (*jwk.Key, error) {
	rawKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	} else if err = rawKey.Validate(); err != nil {
		return nil, err
	}

	key, err := jwk.FromRaw(rawKey)
	if err != nil {
		return nil, err
	}

	err = key.Set(jwk.KeyIDKey, id)
	if err != nil {
		return nil, err
	}

	return &key, nil
}
