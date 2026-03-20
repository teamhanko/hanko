package local_db

import (
	"testing"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v2/test"
)

func TestJWKManagerSuite(t *testing.T) {
	s := new(jwkManagerSuite)
	suite.Run(t, s)
}

type jwkManagerSuite struct {
	test.Suite
}

func (s *jwkManagerSuite) TestDefaultManager() {
	// Test backward compatibility: Multiple encryption keys → Multiple JWKs
	keys := []string{"asfnoadnfoaegnq3094intoaegjnoadjgnoadng", "apdisfoaiegnoaiegnbouaebgn982"}

	persister := s.Storage.GetJwkPersister()

	dm, err := NewDefaultManager(keys, "v1", persister, false)
	require.NoError(s.T(), err)

	// With backward compatibility, 2 encryption keys → 2 JWKs (legacy behavior on fresh DB)
	all, err := persister.GetAll()
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 2, len(all))

	// All should be created as "active" initially (legacy multi-key behavior)
	for _, jwk := range all {
		assert.Equal(s.T(), "active", jwk.State)
		assert.Equal(s.T(), "v1", jwk.EncryptionKeyVersion)
	}

	// GetPublicKeys should return all active JWKs for verification
	js, err := dm.GetPublicKeys()
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 2, js.Len())

	// GetSigningKey should return an active key
	sk, err := dm.GetSigningKey()
	require.NoError(s.T(), err)

	token := jwt.New()
	token.Set("Payload", "isJustFine")
	signed, err := jwt.Sign(token, jwt.WithKey(jwa.RS256, sk))
	require.NoError(s.T(), err)

	// Get Public Key of signing key
	pk, err := sk.PublicKey()
	require.NoError(s.T(), err)

	// Parse and Verify
	tokenParsed, err := jwt.Parse(signed, jwt.WithKey(jwa.RS256, pk))
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), token, tokenParsed)
}

func (s *jwkManagerSuite) TestDefaultManagerSingleKey() {
	// Test with single key (recommended for new deployments)
	keys := []string{"singlekeythatisatleast16chars"}

	persister := s.Storage.GetJwkPersister()

	dm, err := NewDefaultManager(keys, "v1", persister, false)
	require.NoError(s.T(), err)

	// With single key, should create 1 JWK
	all, err := persister.GetAll()
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 1, len(all))
	assert.Equal(s.T(), "active", all[0].State)
	assert.Equal(s.T(), "v1", all[0].EncryptionKeyVersion)

	// GetSigningKey should work
	sk, err := dm.GetSigningKey()
	require.NoError(s.T(), err)
	require.NotNil(s.T(), sk)
}

func (s *jwkManagerSuite) TestKeyRotation() {
	// Test key rotation with existing JWKs
	keys := []string{"testkeythatisatleast16chars"}
	persister := s.Storage.GetJwkPersister()

	// Create initial manager with 1 JWK
	dm, err := NewDefaultManager(keys, "v1", persister, false)
	require.NoError(s.T(), err)

	// Verify initial state
	all, err := persister.GetAll()
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 1, len(all))

	// Rotate the key
	_, err = dm.RotateKey()
	require.NoError(s.T(), err)

	// After rotation, should have 2 keys: 1 active, 1 rotating
	all, err = persister.GetAll()
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 2, len(all))

	// Count states
	activeCount := 0
	rotatingCount := 0
	for _, jwk := range all {
		if jwk.State == "active" {
			activeCount++
		} else if jwk.State == "rotating" {
			rotatingCount++
		}
	}
	assert.Equal(s.T(), 1, activeCount)
	assert.Equal(s.T(), 1, rotatingCount)

	// GetPublicKeys should return both active and rotating
	js, err := dm.GetPublicKeys()
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 2, js.Len())
}
