package local_db

import (
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v2/test"
	"testing"
)

func TestJWKManagerSuite(t *testing.T) {
	s := new(jwkManagerSuite)
	suite.Run(t, s)
}

type jwkManagerSuite struct {
	test.Suite
}

func (s *jwkManagerSuite) TestDefaultManager() {
	keys := []string{"asfnoadnfoaegnq3094intoaegjnoadjgnoadng", "apdisfoaiegnoaiegnbouaebgn982"}

	persister := s.Storage.GetJwkPersister()

	dm, err := NewDefaultManager(keys, persister)
	require.NoError(s.T(), err)
	all, err := persister.GetAll()

	require.NoError(s.T(), err)
	assert.Equal(s.T(), 2, len(all))

	js, err := dm.GetPublicKeys()
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 2, js.Len())

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
