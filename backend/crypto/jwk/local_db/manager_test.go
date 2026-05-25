package local_db

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/test"
)

func TestJWKManagerSuite(t *testing.T) {
	s := new(jwkManagerSuite)
	suite.Run(t, s)
}

type jwkManagerSuite struct {
	test.Suite
}

func (s *jwkManagerSuite) TestDefaultManager() {
	cfg := config.DefaultConfig()
	cfg.SecretKeys = []string{"asfnoadnfoaegnq3094intoaegjnoadjgnoadng", "apdisfoaiegnoaiegnbouaebgn982"}

	persister := s.Storage.GetJwkPersister()

	err := SyncSecretKeys(cfg, s.Storage)
	s.Require().NoError(err)

	dm, err := NewDefaultManager(cfg.SecretKeys, persister)
	require.NoError(s.T(), err)
	all, err := persister.GetAll(uuid.FromStringOrNil(config.DefaultTenantID))

	require.NoError(s.T(), err)
	assert.Equal(s.T(), 2, len(all))

	js, err := dm.GetPublicKeys(uuid.FromStringOrNil(config.DefaultTenantID))
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 2, js.Len())

	sk, err := dm.GetSigningKey(uuid.FromStringOrNil(config.DefaultTenantID))
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
