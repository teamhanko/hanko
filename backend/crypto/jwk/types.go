package jwk

import (
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

// KeyProvider combines all key management capabilities
type KeyProvider interface {
	Manager
	Generator
}

type Manager interface {
	// GenerateKey is used to generate a jwk Key
	GenerateKey() (jwk.Key, error)
	// GetPublicKeys returns all Public keys that are persisted
	GetPublicKeys() (jwk.Set, error)
	// GetSigningKey returns the last added private key that is used for signing
	GetSigningKey() (jwk.Key, error)
}

type Generator interface {
	Sign(jwt.Token) ([]byte, error)
	Verify([]byte) (jwt.Token, error)
}
