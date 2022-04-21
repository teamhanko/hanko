package jwk

import (
	"encoding/json"
	"fmt"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGenerator(t *testing.T) {
	for k, c := range []struct {
		g     KeyGenerator
		use   string
		check func(ks jwk.Key)
	}{
		{
			g:   &RSAKeyGenerator{},
			use: "sig",
			check: func(ks jwk.Key) {
				//assert.Len(t, ks, 2)
				rsaKey, ok := (ks).(jwk.RSAPrivateKey)
				if !ok {
					t.Fail()
				}
				keyId, _ := rsaKey.Get(jwk.KeyIDKey)
				assert.Equal(t, keyId, "my_key_id")
				assert.Equal(t, jwa.RSA, rsaKey.KeyType())
				buf, err := json.MarshalIndent(rsaKey, "", "  ")
				require.NoError(t, err)
				t.Logf("%s\n", buf)
			},
		},
	} {
		t.Run(fmt.Sprintf("case=%d", k), func(t *testing.T) {
			keys, err := c.g.Generate("my_key_id")
			require.NoError(t, err)
			if err == nil {
				c.check(keys)
			}
		})
	}
}
