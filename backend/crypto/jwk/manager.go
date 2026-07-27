package jwk

import (
	"fmt"

	"github.com/teamhanko/hanko/backend/v3/config"
	"github.com/teamhanko/hanko/backend/v3/crypto/jwk/aws_kms"
	"github.com/teamhanko/hanko/backend/v3/crypto/jwk/local_db"
	"github.com/teamhanko/hanko/backend/v3/persistence"
)

func NewManager(cfg config.Config, persister persistence.Persister) (KeyProvider, error) {
	switch cfg.Secrets.KeyManagement.Type {
	case "local":
		return local_db.NewDefaultManager(cfg.SecretKeys, persister.GetJwkPersister())
	case "aws_kms":
		return aws_kms.NewAWSKMSManager(cfg.Secrets.KeyManagement)
	}

	return nil, fmt.Errorf("unsupported key management type: %s", cfg.Secrets.KeyManagement.Type)
}
