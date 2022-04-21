package jwk

import "github.com/lestrrat-go/jwx/v2/jwk"

// KeyGenerator Interface for JSON Web Key Generation
type KeyGenerator interface {
	// Generate a new JWK with a given id
	Generate(id string) (jwk.Key, error)
}
