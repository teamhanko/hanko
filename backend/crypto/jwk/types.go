package jwk

import (
	"github.com/gofrs/uuid"
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
	GenerateKey(tenantID uuid.UUID) (jwk.Key, error)
	// GetPublicKeys returns all Public keys that are persisted
	GetPublicKeys(tenantID uuid.UUID) (jwk.Set, error)
	// GetSigningKey returns the last added private key that is used for signing
	GetSigningKey(tenantID uuid.UUID) (jwk.Key, error)
}

type Generator interface {
	Sign(token jwt.Token, tenantID uuid.UUID) ([]byte, error)
	Verify(tokenString []byte, tenantID uuid.UUID) (jwt.Token, error)
}
