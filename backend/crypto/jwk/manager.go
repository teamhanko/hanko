package jwk

import (
	"fmt"

	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/crypto/jwk/aws_kms"
	"github.com/teamhanko/hanko/backend/v3/crypto/jwk/local_db"
	"github.com/teamhanko/hanko/backend/v3/persistence"
)

func NewManager(cfg config.Secrets, persister persistence.Persister) (KeyProvider, error) {
	switch cfg.KeyManagement.Type {
	case "local":
		return local_db.NewDefaultManager(cfg.Keys, persister.GetJwkPersister())
	case "aws_kms":
		return aws_kms.NewAWSKMSManager(cfg.KeyManagement)
	}

	return nil, fmt.Errorf("unsupported key management type: %s", cfg.KeyManagement.Type)
}
